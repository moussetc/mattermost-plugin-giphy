package main

import (
	"fmt"
	"strings"

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
	err := p.API.RegisterCommand(&model.Command{
		Trigger:          triggerGif,
		Description:      "Post a GIF matching your search",
		DisplayName:      "Giphy Search",
		AutoComplete:     true,
		AutoCompleteDesc: "Post a GIF matching your search",
		AutoCompleteHint: "happy kitty",
	})
	if err != nil {
		return errors.Wrap(err, "Unable to define the following command: "+triggerGif)
	}
	err = p.API.RegisterCommand(&model.Command{
		Trigger:          triggerGifs,
		Description:      "Preview a GIF",
		DisplayName:      "Giphy Shuffle",
		AutoComplete:     true,
		AutoCompleteDesc: "Let you preview and shuffle a GIF before posting for real",
		AutoCompleteHint: "mayhem guy",
	})
	if err != nil {
		return errors.Wrap(err, "Unable to define the following command: "+triggerGifs)
	}
	return nil
}

func getCommandKeywords(commandLine string, trigger string) string {
	return strings.Replace(commandLine, "/"+trigger, "", 1)
}

// executeCommandGif returns a public post containing a matching GIF
func (p *Plugin) executeCommandGif(command string) (*model.CommandResponse, *model.AppError) {
	keywords := getCommandKeywords(command, triggerGif)
	cursor := ""
	gifURL, err := p.gifProvider.GetGifURL(p.getConfiguration(), keywords, &cursor)
	if err != nil {
		return nil, err
	}

	text := generateGifCaption(p.getConfiguration().DisplayMode, keywords, gifURL, p.gifProvider.GetAttributionMessage())
	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, Text: text}, nil
}

// executeCommandGifShuffle returns an ephemeral (private) post with one GIF that can either be posted, shuffled or canceled
func (p *Plugin) executeCommandGifShuffle(command string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cursor := ""
	keywords := getCommandKeywords(command, triggerGifs)
	gifURL, err := p.gifProvider.GetGifURL(p.getConfiguration(), keywords, &cursor)
	if err != nil {
		return nil, err
	}

	post := p.generateGifPost(p.botId, keywords, gifURL, args.ChannelId, args.RootId, p.gifProvider.GetAttributionMessage())
	post.Message = generateGifCaption(p.getConfiguration().DisplayMode, keywords, gifURL, p.gifProvider.GetAttributionMessage())
	post.Props = map[string]interface{}{
		"attachments": generateShufflePostAttachments(keywords, gifURL, cursor, args.RootId),
	}
	p.API.SendEphemeralPost(args.UserId, post)

	return &model.CommandResponse{}, nil
}

func generateGifCaption(displayMode, keywords, gifURL, attributionMessage string) string {
	if displayMode == "full_url" {
		return fmt.Sprintf("**/gif [%s](%s)** : %s *%s*", keywords, gifURL, gifURL, attributionMessage)
	}
	return fmt.Sprintf("**/gif [%s](%s)** \n\n*%s* \n\n![GIF for '%s'](%s)", keywords, gifURL, attributionMessage, keywords, gifURL)
}

func (p *Plugin) generateGifPost(userId, keywords, gifURL, channelId, rootId, attributionMessage string) *model.Post {
	return &model.Post{
		Message:   generateGifCaption(p.getConfiguration().DisplayMode, keywords, gifURL, attributionMessage),
		UserId:    userId,
		ChannelId: channelId,
		RootId:    rootId,
	}
}

func generateShufflePostAttachments(keywords, gifURL, cursor, rootId string) []*model.SlackAttachment {
	actionContext := map[string]interface{}{
		contextKeywords: keywords,
		contextGifURL:   gifURL,
		contextCursor:   cursor,
		contextRootId:   rootId,
	}

	actions := []*model.PostAction{}
	actions = append(actions, generateButton("Cancel", URLCancel, actionContext))
	actions = append(actions, generateButton("Shuffle", URLShuffle, actionContext))
	actions = append(actions, generateButton("Send", URLSend, actionContext))

	attachments := []*model.SlackAttachment{}
	attachments = append(attachments, &model.SlackAttachment{
		Actions: actions,
	})

	return attachments
}

// Generate an attachment for an action Button that will point to a plugin HTTP handler
func generateButton(name string, urlAction string, context map[string]interface{}) *model.PostAction {
	return &model.PostAction{
		Name: name,
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL:     fmt.Sprintf("/plugins/%s%s", manifest.Id, urlAction),
			Context: context,
		},
	}
}
