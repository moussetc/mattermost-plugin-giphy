package main

import (
	"errors"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mitchellh/mapstructure"
)

// Contains what's related to handling HTTP requests directed to the plugin

const (
	URLShuffle = "/shuffle"
	URLCancel  = "/cancel"
	URLSend    = "/send"
)

type integrationRequest struct {
	Keywords string `mapstructure:"keywords"`
	GifURL   string `mapstructure:"gifURL"`
	Cursor   string `mapstructure:"cursor"`
	RootId   string `mapstructure:"rootId"`
	model.PostActionIntegrationRequest
}

type (
	pluginHTTPHandler interface {
		handleCancel(p *Plugin, w http.ResponseWriter, request *integrationRequest)
		handleShuffle(p *Plugin, w http.ResponseWriter, request *integrationRequest)
		handleSend(p *Plugin, w http.ResponseWriter, request *integrationRequest)
	}
	defaultHTTPHandler struct{}
)

var postActionIntegrationRequestFromJson = model.PostActionIntegrationRequestFromJson

var notifyHandlerError = defaultNotifyHandlerError

func (p *Plugin) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Header is set by MM server only if the request was successfully authenticated
	userId := r.Header.Get("Mattermost-User-Id")
	if userId == "" {
		http.Error(w, "Authentication failed: user not set in header", http.StatusUnauthorized)
		return
	}

	request, err := parseRequest(r)
	if err != nil {
		p.API.LogWarn("Could not parse PostActionIntegrationRequest: "+err.Error(), nil)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userId != request.UserId {
		http.Error(w, "The user of the request should match the authenticated user", http.StatusBadRequest)
		return
	}
	if !p.API.HasPermissionToChannel(request.UserId, request.ChannelId, model.PERMISSION_READ_CHANNEL) {
		http.Error(w, "The user is not allowed to read this channel", http.StatusForbidden)
		return
	}
	if !p.API.HasPermissionToChannel(request.UserId, request.ChannelId, model.PERMISSION_CREATE_POST) {
		http.Error(w, "The user is not allowed to post in this channel", http.StatusForbidden)
		return
	}

	switch r.URL.Path {
	case URLShuffle:
		p.httpHandler.handleShuffle(p, w, request)
	case URLSend:
		p.httpHandler.handleSend(p, w, request)
	case URLCancel:
		p.httpHandler.handleCancel(p, w, request)
	default:
		http.NotFound(w, r)
	}
}

func parseRequest(r *http.Request) (*integrationRequest, error) {
	// Read data added by default for a button action
	request := postActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		return nil, errors.New("request cannot be nil")
	}

	context := integrationRequest{}
	context.PostActionIntegrationRequest = *request
	err := mapstructure.Decode(request.Context, &context)
	if context.Keywords == "" {
		return nil, errors.New("Missing " + contextKeywords + " from action request context")
	}
	if context.GifURL == "" {
		return nil, errors.New("Missing " + contextGifURL + " from action request context")
	}
	if context.Cursor == "" {
		return nil, errors.New("Missing " + contextCursor + " from action request context")
	}
	if context.RootId == "" {
		return nil, errors.New("Missing " + contextRootId + " from action request context")
	}
	return &context, err
}

func writeResponse(httpStatus int, w http.ResponseWriter) {
	w.WriteHeader(httpStatus)
	if httpStatus == http.StatusOK {
		// Return the object the MM server expects in case of 200 status
		response := &model.PostActionIntegrationResponse{}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response.ToJson())
	}
}

// Delete the ephemeral shuffle post
func (h *defaultHTTPHandler) handleCancel(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	p.API.DeleteEphemeralPost(request.UserId, request.PostId)
	writeResponse(http.StatusOK, w)
}

// Replace the GIF in the ephemeral shuffle post by a new one
func (h *defaultHTTPHandler) handleShuffle(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	shuffledGifURL, err := p.gifProvider.GetGifURL(request.Keywords, &request.Cursor)
	if err != nil {
		notifyHandlerError(p.API, "Unable to fetch a new Gif for shuffling", err, &request.PostActionIntegrationRequest)
		writeResponse(http.StatusServiceUnavailable, w)
		return
	}

	post := &model.Post{
		Id:        request.PostId,
		ChannelId: request.ChannelId,
		UserId:    p.botId,
		RootId:    request.RootId,
		Message:   generateGifCaption(p.getConfiguration().DisplayMode, request.Keywords, shuffledGifURL, p.gifProvider.GetAttributionMessage()),
		Props: map[string]interface{}{
			"attachments": generateShufflePostAttachments(request.Keywords, shuffledGifURL, request.Cursor, request.RootId),
		},
		CreateAt: model.GetMillis(),
		UpdateAt: model.GetMillis(),
	}

	p.API.UpdateEphemeralPost(request.UserId, post)
	writeResponse(http.StatusOK, w)
}

// Post the actual GIF and delete the obsolete ephemeral post
func (h *defaultHTTPHandler) handleSend(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	p.API.DeleteEphemeralPost(request.UserId, request.PostId)
	post := &model.Post{
		Message:   generateGifCaption(p.getConfiguration().DisplayMode, request.Keywords, request.GifURL, p.gifProvider.GetAttributionMessage()),
		UserId:    request.UserId,
		ChannelId: request.ChannelId,
		RootId:    request.RootId,
	}
	_, err := p.API.CreatePost(post)
	if err != nil {
		notifyHandlerError(p.API, "Unable to create post : ", err, &request.PostActionIntegrationRequest)
		writeResponse(http.StatusInternalServerError, w)
		return
	}

	writeResponse(http.StatusOK, w)
}

// Informs the user of an error that occured in a button handler (no direct response possible so it use ephemeral messages), and also logs it
func defaultNotifyHandlerError(api plugin.API, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
	fullMessage := manifest.Name + ":"
	if err != nil {
		fullMessage = err.Message
	} else {
		fullMessage = message
	}
	api.SendEphemeralPost(request.UserId, &model.Post{
		Message:   "*" + fullMessage + "*",
		ChannelId: request.ChannelId,
		Props: map[string]interface{}{
			"sent_by_plugin": true,
		},
	})
	api.LogWarn(message, err)
}
