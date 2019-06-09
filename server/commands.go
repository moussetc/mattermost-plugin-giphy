package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/model"

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
	gifURL, err := p.gifProvider.getGifURL(p.getConfiguration(), keywords, &cursor)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	text := generateGifCaption(keywords, gifURL)
	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, Text: text}, nil
}

// executeCommandGifShuffle returns an ephemeral (private) post with one GIF that can either be posted, shuffled or canceled
func (p *Plugin) executeCommandGifShuffle(command string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cursor := ""
	keywords := getCommandKeywords(command, triggerGifs)
	gifURL, err := p.gifProvider.getGifURL(p.getConfiguration(), keywords, &cursor)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	text := generateGifCaption(keywords, gifURL)
	attachments := generateShufflePostAttachments(p, keywords, gifURL, cursor)

	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL, Text: text, Attachments: attachments}, nil
}

func generateGifCaption(keywords string, gifURL string) string {
	return " */gif [" + keywords + "](" + gifURL + ")*\n" + "![GIF for '" + keywords + "'](" + gifURL + ")"
}

func generateShufflePostAttachments(p *Plugin, keywords string, gifURL string, cursor string) []*model.SlackAttachment {
	actionContext := map[string]interface{}{
		contextKeywords: keywords,
		contextGifURL:   gifURL,
		contextCursor:   cursor,
	}

	actions := []*model.PostAction{}
	actions = append(actions, generateButton(p.siteURL, "Cancel", URLCancel, actionContext))
	actions = append(actions, generateButton(p.siteURL, "Shuffle", URLShuffle, actionContext))
	actions = append(actions, generateButton(p.siteURL, "Send", URLSend, actionContext))

	attachments := []*model.SlackAttachment{}
	attachments = append(attachments, &model.SlackAttachment{
		Actions: actions,
	})

	return attachments
}

// Generate an attachment for an action Button that will point to a plugin HTTP handler
func generateButton(siteURL string, name string, urlAction string, context map[string]interface{}) *model.PostAction {
	return &model.PostAction{
		Name: name,
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL:     fmt.Sprintf("%s/plugins/%s"+urlAction, siteURL, manifest.Id),
			Context: context,
		},
	}
}
