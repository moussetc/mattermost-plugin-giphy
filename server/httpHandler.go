package main

import (
	"errors"
	"net/http"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"

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
	Caption  string `mapstructure:"caption"`
	GifURL   string `mapstructure:"gifURL"`
	Cursor   string `mapstructure:"cursor"`
	RootID   string `mapstructure:"rootID"`
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

var notifyUserOfError = defaultNotifyUserOfError

func (p *Plugin) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Header is set by MM server only if the request was successfully authenticated
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Authentication failed: user not set in header", http.StatusUnauthorized)
		return
	}

	request, err := parseRequest(r)
	if err != nil {
		p.API.LogWarn("Could not parse PostActionIntegrationRequest: "+err.Error(), nil)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userID != request.UserId {
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
		return nil, errors.New("missing " + contextKeywords + " from action request context")
	}
	if context.GifURL == "" {
		return nil, errors.New("missing " + contextGifURL + " from action request context")
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
	if request.Cursor == "" {
		notifyUserOfError(p.API, p.botID, "No more GIFs found for '"+request.Keywords+"'", nil, &request.PostActionIntegrationRequest)
		return
	}
	shuffledGifURL, err := p.gifProvider.GetGifURL(request.Keywords, &request.Cursor)
	if err != nil {
		notifyUserOfError(p.API, p.botID, "Unable to fetch a new Gif for shuffling", err, &request.PostActionIntegrationRequest)
		writeResponse(http.StatusServiceUnavailable, w)
		return
	}
	if shuffledGifURL == "" {
		notifyUserOfError(p.API, p.botID, "No GIFs found for '"+request.Keywords+"'", nil, &request.PostActionIntegrationRequest)
		return
	}
	post := &model.Post{
		Id:        request.PostId,
		ChannelId: request.ChannelId,
		UserId:    p.botID,
		RootId:    request.RootID,
		// Only embedded display mode works inside an ephemeral post
		Message:  generateGifCaption(pluginConf.DisplayModeEmbedded, request.Keywords, request.Caption, shuffledGifURL, p.gifProvider.GetAttributionMessage()),
		CreateAt: model.GetMillis(),
		UpdateAt: model.GetMillis(),
	}
	post.SetProps(map[string]interface{}{
		"attachments": generateShufflePostAttachments(request.Keywords, request.Caption, shuffledGifURL, request.Cursor, request.RootID),
	})
	p.API.UpdateEphemeralPost(request.UserId, post)
	writeResponse(http.StatusOK, w)
}

// Post the actual GIF and delete the obsolete ephemeral post
func (h *defaultHTTPHandler) handleSend(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	p.API.DeleteEphemeralPost(request.UserId, request.PostId)
	post := &model.Post{
		Message:   generateGifCaption(p.getConfiguration().DisplayMode, request.Keywords, request.Caption, request.GifURL, p.gifProvider.GetAttributionMessage()),
		UserId:    request.UserId,
		ChannelId: request.ChannelId,
		RootId:    request.RootID,
	}
	_, err := p.API.CreatePost(post)
	if err != nil {
		notifyUserOfError(p.API, p.botID, "Unable to create post : ", err, &request.PostActionIntegrationRequest)
		writeResponse(http.StatusInternalServerError, w)
		return
	}

	writeResponse(http.StatusOK, w)
}

// Informs the user of an error (domain error with message, or technical error with err) that occurred in a button handler, and logs it if it's technical
func defaultNotifyUserOfError(api plugin.API, botID string, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
	fullMessage := message
	if err != nil {
		fullMessage = err.Message
	}
	post := &model.Post{
		Message:   "*" + fullMessage + "*",
		ChannelId: request.ChannelId,
		UserId:    botID,
	}
	post.SetProps(map[string]interface{}{
		"sent_by_plugin": true,
	})
	api.SendEphemeralPost(request.UserId, post)

	// Only log technical errors
	if err != nil {
		api.LogWarn(message, err)
	}
}
