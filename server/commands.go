package main

import (
	"fmt"
	"regexp"
	"strings"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/pkg/errors"
)

// Contains all that's related to the basic Post command

// Triggers used to define slash commands
const (
	triggerGif  = "gif"
	triggerGifs = "gifs"
)

func (p *Plugin) RegisterCommands() error {
	unregisterErr := p.API.UnregisterCommand("", triggerGif)
	if unregisterErr != nil {
		p.API.LogWarn("Unable to unregister the " + triggerGif + " command" + unregisterErr.Error())
	}
	unregisterErr = p.API.UnregisterCommand("", triggerGifs)
	if unregisterErr != nil {
		p.API.LogWarn("Unable to unregister the " + triggerGifs + " command" + unregisterErr.Error())
	}

	config := p.getConfiguration()
	if config.CommandTriggerGif != "" {
		err := p.API.RegisterCommand(&model.Command{
			Trigger:          config.CommandTriggerGif,
			Description:      "Post a GIF matching your search",
			DisplayName:      "Giphy Search",
			AutoComplete:     true,
			AutoCompleteDesc: "Post a GIF matching your search",
			AutoCompleteHint: getHintMessage(config.CommandTriggerGif),
		})
		if err != nil {
			return errors.Wrap(err, "Unable to define the following command: "+config.CommandTriggerGif)
		}
	}
	if config.CommandTriggerGifWithPreview != "" {
		err := p.API.RegisterCommand(&model.Command{
			Trigger:          config.CommandTriggerGifWithPreview,
			Description:      "Preview a GIF",
			DisplayName:      "Giphy Shuffle",
			AutoComplete:     true,
			AutoCompleteDesc: "Let you preview and shuffle a GIF before posting for real",
			AutoCompleteHint: getHintMessage(config.CommandTriggerGifWithPreview),
		})
		if err != nil {
			return errors.Wrap(err, "Unable to define the following command: "+config.CommandTriggerGifWithPreview)
		}
	}
	return nil
}

func parseCommandLine(commandLine, trigger string) (keywords, caption string, err error) {
	reg, err := regexp.Compile("^\\s*(?P<keywords>(\"([^\\s\"]+\\s*)+\")+|([^\\s\"]+\\s*)+)(?P<caption>\\s+\"(\\s*[^\\s\"]+\\s*)+\")?\\s*$")
	if err != nil {
		return "", "", errors.New("Could not compile regexp")
	}
	matchIndexes := reg.FindStringSubmatch(strings.Replace(commandLine, "/"+trigger, "", 1))
	if matchIndexes == nil {
		return "", "", errors.New(fmt.Sprintf("Could not read the command, try one of the following syntax: /%s %s", trigger, getHintMessage(trigger)))
	}
	results := make(map[string]string)
	for i, name := range reg.SubexpNames() {
		results[name] = matchIndexes[i]
	}
	return strings.Trim(strings.TrimSpace(results["keywords"]), "\""), strings.Trim(strings.TrimSpace(results["caption"]), "\""), nil
}

// executeCommandGif returns a public post containing a matching GIF
func (p *Plugin) executeCommandGif(keywords, caption string) (*model.CommandResponse, *model.AppError) {
	cursor := ""
	gifURL, errGif := p.gifProvider.GetGifURL(keywords, &cursor)
	if errGif != nil {
		return nil, errGif
	}

	text := generateGifCaption(p.getConfiguration().DisplayMode, keywords, caption, gifURL, p.gifProvider.GetAttributionMessage())
	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, Text: text}, nil
}

// executeCommandGifWithPreview returns an ephemeral post with one GIF that can either be posted, shuffled or canceled
func (p *Plugin) executeCommandGifWithPreview(keywords, caption string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cursor := ""
	gifURL, errGif := p.gifProvider.GetGifURL(keywords, &cursor)
	if errGif != nil {
		return nil, errGif
	}

	post := p.generateGifPost(p.botId, keywords, caption, gifURL, args.ChannelId, args.RootId, p.gifProvider.GetAttributionMessage())
	// Only embedded display mode works inside an ephemeral post
	post.Message = generateGifCaption(pluginConf.DisplayModeEmbedded, keywords, caption, gifURL, p.gifProvider.GetAttributionMessage())
	post.Props = map[string]interface{}{
		"attachments": generateShufflePostAttachments(keywords, caption, gifURL, cursor, args.RootId),
	}
	p.API.SendEphemeralPost(args.UserId, post)

	return &model.CommandResponse{}, nil
}

func getHintMessage(trigger string) string {
	return "[happy kitty] or /" + trigger + " \"[happy kitty]\" \"[This is a custom caption]\""
}

func generateGifCaption(displayMode, keywords, caption, gifURL, attributionMessage string) string {
	captionOrKeywords := caption
	if caption == "" {
		captionOrKeywords = fmt.Sprintf("/gif **%s**", keywords)
	}
	if displayMode == pluginConf.DisplayModeFullURL {
		return fmt.Sprintf("%s \n\n%s *%s*", captionOrKeywords, gifURL, attributionMessage)
	}
	return fmt.Sprintf("%s \n\n*%s* \n\n![GIF for '%s'](%s)", captionOrKeywords, attributionMessage, keywords, gifURL)
}

func (p *Plugin) generateGifPost(userId, keywords, caption, gifURL, channelId, rootId, attributionMessage string) *model.Post {
	return &model.Post{
		Message:   generateGifCaption(p.getConfiguration().DisplayMode, keywords, caption, gifURL, attributionMessage),
		UserId:    userId,
		ChannelId: channelId,
		RootId:    rootId,
	}
}

func generateShufflePostAttachments(keywords, caption, gifURL, cursor, rootId string) []*model.SlackAttachment {
	actionContext := map[string]interface{}{
		contextKeywords: keywords,
		contextCaption:  caption,
		contextGifURL:   gifURL,
		contextCursor:   cursor,
		contextRootId:   rootId,
	}

	actions := []*model.PostAction{}
	actions = append(actions, generateButton("Cancel", URLCancel, "default", actionContext))
	actions = append(actions, generateButton("Shuffle", URLShuffle, "primary", actionContext))
	actions = append(actions, generateButton("Send", URLSend, "good", actionContext))

	attachments := []*model.SlackAttachment{}
	attachments = append(attachments, &model.SlackAttachment{
		Actions: actions,
	})

	return attachments
}

// Generate an attachment for an action Button that will point to a plugin HTTP handler
func generateButton(name string, urlAction string, style string, context map[string]interface{}) *model.PostAction {
	return &model.PostAction{
		Name:  name,
		Type:  model.POST_ACTION_TYPE_BUTTON,
		Style: style,
		Integration: &model.PostActionIntegration{
			URL:     fmt.Sprintf("/plugins/%s%s", manifest.Id, urlAction),
			Context: context,
		},
	}
}
