package main

import (
	"github.com/mattermost/mattermost-server/model"
)

// Contains all that's related to the basic Post command

// executeCommandGif returns a public post containing a matching GIF
func (p *GiphyPlugin) executeCommandGif(command string) (*model.CommandResponse, *model.AppError) {
	keywords := getCommandKeywords(command, triggerGif)
	gifURL, err := p.gifProvider.getGifURL(p.config(), keywords, 0)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	text := p.generateGifCaption(keywords, gifURL)
	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, Text: text}, nil
}

func (p *GiphyPlugin) generateGifCaption(keywords string, gifURL string) string {
	return " *[" + keywords + "](" + gifURL + ")*\n" + "![GIF for '" + keywords + "'](" + gifURL + ")"
}
