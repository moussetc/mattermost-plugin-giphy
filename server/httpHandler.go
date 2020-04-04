package main

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Contains what's related to handling HTTP requests directed to the plugin

const (
	URLShuffle = "/shuffle"
	URLCancel  = "/cancel"
	URLSend    = "/send"
)

type (
	pluginHTTPHandler interface {
		handleCancel(p *Plugin, w http.ResponseWriter, r *http.Request)
		handleShuffle(p *Plugin, w http.ResponseWriter, r *http.Request)
		handlePost(p *Plugin, w http.ResponseWriter, r *http.Request)
	}
	defaultHTTPHandler struct{}
)

var postActionIntegrationRequestFromJson = model.PostActionIntegrationRequestFromJson

var notifyHandlerError = defaultNotifyHandlerError

func (p *Plugin) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case URLShuffle:
		p.httpHandler.handleShuffle(p, w, r)
	case URLSend:
		p.httpHandler.handlePost(p, w, r)
	case URLCancel:
		p.httpHandler.handleCancel(p, w, r)
	default:
		http.NotFound(w, r)
	}
}

func parseRequest(api plugin.API, w http.ResponseWriter, r *http.Request) (request *model.PostActionIntegrationRequest, gifURL string, keywords string, cursor string, ok bool) {
	// Read data added by default for a button action
	request = postActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		api.LogWarn("Could not parse PostActionIntegrationRequest", nil)
		w.WriteHeader(http.StatusBadRequest)
		return nil, "", "", "", false
	}
	gifURLObj, ok := request.Context[contextGifURL]
	if !ok {
		notifyHandlerError(api, "Missing "+contextGifURL+" from action request context", nil, request)
		w.WriteHeader(http.StatusBadRequest)
		return nil, "", "", "", false
	}
	gifURL = gifURLObj.(string)

	keywordsObj, ok := request.Context[contextKeywords]
	if !ok {
		notifyHandlerError(api, "Missing "+contextKeywords+" from action request context", nil, request)
		w.WriteHeader(http.StatusBadRequest)
		return nil, "", "", "", false
	}
	keywords = keywordsObj.(string)

	cursorObj, ok := request.Context[contextCursor]
	if !ok {
		notifyHandlerError(api, "Missing "+contextCursor+" from action request context", nil, request)
		w.WriteHeader(http.StatusBadRequest)
		return nil, "", "", "", false
	}
	cursor = cursorObj.(string)

	return request, gifURL, keywords, cursor, true
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

// handleCancel delete the ephemeral shuffle post
func (h *defaultHTTPHandler) handleCancel(p *Plugin, w http.ResponseWriter, r *http.Request) {
	request, _, _, _, ok := parseRequest(p.API, w, r)
	if !ok {
		return
	}

	p.API.DeleteEphemeralPost(request.UserId, request.PostId)

	writeResponse(http.StatusOK, w)
}

// handleShuffle replace the GIF in the ephemeral shuffle post by a new one
func (h *defaultHTTPHandler) handleShuffle(p *Plugin, w http.ResponseWriter, r *http.Request) {
	request, _, keywords, cursor, ok := parseRequest(p.API, w, r)
	if !ok {
		return
	}

	shuffledGifURL, err := p.gifProvider.getGifURL(p.getConfiguration(), keywords, &cursor)
	if err != nil {
		notifyHandlerError(p.API, "Unable to fetch a new Gif for shuffling", err, request)
		writeResponse(http.StatusServiceUnavailable, w)
		return
	}

	post := &model.Post{
		Id:        request.PostId,
		ChannelId: request.ChannelId,
		UserId:    request.UserId,
		Message:   generateGifCaption(keywords, shuffledGifURL),
		Props: map[string]interface{}{
			"attachments": generateShufflePostAttachments(p, keywords, shuffledGifURL, cursor),
		},
		CreateAt: model.GetMillis(),
		UpdateAt: model.GetMillis(),
	}

	p.API.UpdateEphemeralPost(request.UserId, post)
	writeResponse(http.StatusOK, w)
}

// handlePost post the actual GIF and delete the obsolete ephemeral post
func (h *defaultHTTPHandler) handlePost(p *Plugin, w http.ResponseWriter, r *http.Request) {
	request, gifURL, keywords, _, ok := parseRequest(p.API, w, r)
	if !ok {
		return
	}

	p.API.DeleteEphemeralPost(request.UserId, request.PostId)
	post := &model.Post{
		Message:   generateGifCaption(keywords, gifURL),
		UserId:    request.UserId,
		ChannelId: request.ChannelId,
	}
	_, err := p.API.CreatePost(post)
	if err != nil {
		notifyHandlerError(p.API, "Unable to create post : ", err, request)
		writeResponse(http.StatusInternalServerError, w)
		return
	}

	writeResponse(http.StatusOK, w)
}

// notifyHandlerError informs the user of an error that occured in a buttion handler (no direct response possible so it use ephemeral messages), and also logs it
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
