package main

import (
	"net/http"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

// Contains what's related to the Shuffle command

// executeCommandGifShuffle returns an ephemeral (private) post with one GIF that can either be posted, shuffled or canceled
func (p *Plugin) executeCommandGifShuffle(command string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cursor := ""
	keywords := getCommandKeywords(command, triggerGifs)
	gifURL, err := p.gifProvider.getGifURL(&p.API, p.getConfiguration(), keywords, &cursor)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	text := p.generateGifCaption(keywords, gifURL)
	attachments := p.generateShufflePostAttachments(keywords, gifURL, cursor)

	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL, Text: text, Attachments: attachments}, nil
}

type handlerFunc func(request *model.PostActionIntegrationRequest, keywords string, gifURL string, cursor string) int

var postActionIntegrationRequestFromJson = model.PostActionIntegrationRequestFromJson

// handleHTTPAction reads the Gif context for an action (buttons) and execute the action
func (p *Plugin) handleHTTPAction(action handlerFunc, c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	// Read data added by default for a button action
	request := postActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	gifURL, ok := request.Context[contextGifURL]
	if !ok {
		p.logHandlerError("Missing "+contextGifURL+" from action request context", nil, request)
		w.WriteHeader(http.StatusBadRequest)
	}
	keywords, ok := request.Context[contextKeywords]
	if !ok {
		p.logHandlerError("Missing "+contextKeywords+" from action request context", nil, request)
		w.WriteHeader(http.StatusBadRequest)
	}
	cursor, ok := request.Context[contextCursor]
	if !ok {
		p.logHandlerError("Missing "+contextCursor+" from action request context", nil, request)
		w.WriteHeader(http.StatusBadRequest)
	}

	httpStatus := action(request, keywords.(string), gifURL.(string), cursor.(string))
	w.WriteHeader(httpStatus)
	if httpStatus == http.StatusOK {
		// Return the object the MM server expects in case of 200 status
		response := &model.PostActionIntegrationResponse{}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response.ToJson())
	}
}

func (p *Plugin) generateShufflePostAttachments(keywords string, gifURL string, cursor string) []*model.SlackAttachment {
	actionContext := map[string]interface{}{
		contextKeywords: keywords,
		contextGifURL:   gifURL,
		contextCursor:   cursor,
	}

	actions := []*model.PostAction{}
	actions = append(actions, p.generateButton("Cancel", URLCancel, actionContext))
	actions = append(actions, p.generateButton("Shuffle", URLShuffle, actionContext))
	actions = append(actions, p.generateButton("Send", URLSend, actionContext))

	attachments := []*model.SlackAttachment{}
	attachments = append(attachments, &model.SlackAttachment{
		Actions: actions,
	})

	return attachments
}

// handleCancel delete the ephemeral shuffle post
func (p *Plugin) handleCancel(request *model.PostActionIntegrationRequest, keywords string, gifURL string, cursor string) int {
	post := &model.Post{
		Id: request.PostId,
	}
	p.API.DeleteEphemeralPost(request.UserId, post)

	return http.StatusOK
}

// handleShuffle replace the GIF in the ephemeral shuffle post by a new one
func (p *Plugin) handleShuffle(request *model.PostActionIntegrationRequest, keywords string, gifURL string, cursor string) int {
	shuffledGifURL, err := p.gifProvider.getGifURL(&p.API, p.getConfiguration(), keywords, &cursor)
	if err != nil {
		p.logHandlerError("Unable to fetch a new Gif for shuffling", err, request)
		return http.StatusServiceUnavailable
	}

	post := &model.Post{
		Id:        request.PostId,
		ChannelId: request.ChannelId,
		UserId:    request.UserId,
		Message:   p.generateGifCaption(keywords, shuffledGifURL),
		Props: map[string]interface{}{
			"attachments": p.generateShufflePostAttachments(keywords, shuffledGifURL, cursor),
		},
		CreateAt: model.GetMillis(),
		UpdateAt: model.GetMillis(),
	}
	p.API.UpdateEphemeralPost(request.UserId, post)
	return http.StatusOK
}

// handlePost post the actual GIF and delete the obsolete ephemeral post
func (p *Plugin) handlePost(request *model.PostActionIntegrationRequest, keywords string, gifURL string, cursor string) int {
	ephemeralPost := &model.Post{
		Id: request.PostId,
	}
	p.API.DeleteEphemeralPost(request.UserId, ephemeralPost)
	post := &model.Post{
		Message:   p.generateGifCaption(keywords, gifURL),
		UserId:    request.UserId,
		ChannelId: request.ChannelId,
	}
	_, err := p.API.CreatePost(post)
	if err != nil {
		p.logHandlerError("Unable to create post : ", err, request)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// logHandlerError informs the user of an error that occured in a buttion handler, and also logs it
func (p *Plugin) logHandlerError(message string, err error, request *model.PostActionIntegrationRequest) {
	fullMessage := "Giphy Plugin: " + message
	if err != nil {
		fullMessage = fullMessage + "\n`" + err.Error() + "`"
	}
	p.API.SendEphemeralPost(request.UserId, &model.Post{
		Message:   fullMessage,
		ChannelId: request.ChannelId,
		Props: map[string]interface{}{
			"sent_by_plugin": true,
		},
	})
	p.API.LogWarn(message, appError("", err))
}
