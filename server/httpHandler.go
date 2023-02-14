package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mitchellh/mapstructure"
)

// Contains what's related to handling HTTP requests directed to the plugin

const (
	URLShuffle  = "/shuffle"
	URLCancel   = "/cancel"
	URLPrevious = "/previous"
	URLSend     = "/send"
)

type integrationRequest struct {
	Keywords        string   `mapstructure:"keywords"`
	Caption         string   `mapstructure:"caption"`
	GifURLs         []string `mapstructure:"gifURLs"`
	CurrentGifIndex int      `mapstructure:"currentGifIndex"`
	SearchCursor    string   `mapstructure:"searchCursor"`
	RootID          string   `mapstructure:"rootID"`
	model.PostActionIntegrationRequest
}

type (
	pluginHTTPHandler interface {
		handleCancel(p *Plugin, w http.ResponseWriter, request *integrationRequest)
		handleShuffle(p *Plugin, w http.ResponseWriter, request *integrationRequest)
		handlePrevious(p *Plugin, w http.ResponseWriter, request *integrationRequest)
		handleSend(p *Plugin, w http.ResponseWriter, request *integrationRequest)
	}
	defaultHTTPHandler struct{}
)

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
		p.API.LogWarn("Could not parse PostActionIntegrationRequest", "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userID != request.UserId {
		http.Error(w, "The user of the request should match the authenticated user", http.StatusBadRequest)
		return
	}
	if !p.API.HasPermissionToChannel(request.UserId, request.ChannelId, model.PermissionReadChannel) {
		http.Error(w, "The user is not allowed to read this channel", http.StatusForbidden)
		return
	}
	if !p.API.HasPermissionToChannel(request.UserId, request.ChannelId, model.PermissionCreatePost) {
		http.Error(w, "The user is not allowed to post in this channel", http.StatusForbidden)
		return
	}

	switch r.URL.Path {
	case URLShuffle:
		p.httpHandler.handleShuffle(p, w, request)
	case URLPrevious:
		p.httpHandler.handlePrevious(p, w, request)
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
	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		return nil, readErr
	}
	var request model.PostActionIntegrationRequest
	jsonErr := json.Unmarshal(body, &request)
	if jsonErr != nil {
		return nil, jsonErr
	}

	context := integrationRequest{}
	context.PostActionIntegrationRequest = request
	err := mapstructure.Decode(request.Context, &context)
	if context.Keywords == "" {
		return nil, errors.New("missing " + contextKeywords + " from action request context")
	}
	if len(context.GifURLs) == 0 {
		return nil, errors.New("missing " + contextGifURLs + " from action request context")
	}
	return &context, err
}

func writeResponse(httpStatus int, w http.ResponseWriter) {
	w.WriteHeader(httpStatus)
	if httpStatus == http.StatusOK {
		// Return the object the MM server expects in case of 200 status
		response := &model.PostActionIntegrationResponse{}
		w.Header().Set("Content-Type", "application/json")
		json, jsonErr := json.Marshal(response)
		if jsonErr == nil {
			_, _ = w.Write(json)
		}
	}
}

// Delete the ephemeral preview post
func (h *defaultHTTPHandler) handleCancel(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	p.API.DeleteEphemeralPost(request.UserId, request.PostId)
	writeResponse(http.StatusOK, w)
}

// Replace the GIF in the ephemeral shuffle post by a new one
func (h *defaultHTTPHandler) handleShuffle(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	random := p.configuration.RandomSearch
	if !random && request.SearchCursor == "" {
		notifyUserOfError(p.API, p.botID, "No more GIFs found for '"+request.Keywords+"'", nil, &request.PostActionIntegrationRequest)
		return
	}
	shuffledGifURL, err := p.gifProvider.GetGifURL(request.Keywords, &request.SearchCursor, random)
	if err != nil {
		notifyUserOfError(p.API, p.botID, "Unable to fetch a new Gif for shuffling", err, &request.PostActionIntegrationRequest)
		writeResponse(http.StatusServiceUnavailable, w)
		return
	}
	if shuffledGifURL == "" {
		notifyUserOfError(p.API, p.botID, "No GIFs found for '"+request.Keywords+"'", nil, &request.PostActionIntegrationRequest)
		return
	}

	gifURLs := append(request.GifURLs, shuffledGifURL)
	h.sendPreviewPost(p, w, request, gifURLs, len(gifURLs)-1)
}

// Replace the GIF in the ephemeral shuffle post by one that was already shuffled
func (h *defaultHTTPHandler) handlePrevious(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	previousIndex := request.CurrentGifIndex - 1
	if previousIndex < 0 {
		notifyUserOfError(p.API, p.botID, "There is no previous URL", nil, &request.PostActionIntegrationRequest)
		writeResponse(http.StatusBadRequest, w)
		return
	}

	h.sendPreviewPost(p, w, request, request.GifURLs, previousIndex)
}

// Create and send an ephemeral for a gif preview message
func (h *defaultHTTPHandler) sendPreviewPost(p *Plugin, w http.ResponseWriter, request *integrationRequest, gifURLs []string, currentGifIndex int) {
	time := model.GetMillis()
	post := &model.Post{
		Id:        request.PostId,
		ChannelId: request.ChannelId,
		UserId:    p.botID,
		RootId:    request.RootID,
		// Only embedded display mode works inside an ephemeral post
		Message:  generateGifCaption(pluginConf.DisplayModeEmbedded, request.Keywords, request.Caption, gifURLs[currentGifIndex], p.gifProvider.GetAttributionMessage()),
		CreateAt: time,
		UpdateAt: time,
	}
	post.SetProps(map[string]interface{}{
		"attachments": generatePreviewPostAttachments(request.Keywords, request.Caption, request.SearchCursor, request.RootID, gifURLs, currentGifIndex),
	})
	p.API.UpdateEphemeralPost(request.UserId, post)
	writeResponse(http.StatusOK, w)
}

// Post the actual GIF and delete the obsolete ephemeral post
func (h *defaultHTTPHandler) handleSend(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	p.API.DeleteEphemeralPost(request.UserId, request.PostId)
	if request.CurrentGifIndex < 0 || request.CurrentGifIndex >= len(request.GifURLs) {
		notifyUserOfError(p.API, p.botID, "Unable to create post : index "+strconv.Itoa(request.CurrentGifIndex)+"is out of bounds [0,"+strconv.Itoa(len(request.GifURLs))+"]", nil, &request.PostActionIntegrationRequest)
		writeResponse(http.StatusInternalServerError, w)
		return
	}
	time := model.GetMillis()
	post := &model.Post{
		Message:   generateGifCaption(p.getConfiguration().DisplayMode, request.Keywords, request.Caption, request.GifURLs[request.CurrentGifIndex], p.gifProvider.GetAttributionMessage()),
		UserId:    request.UserId,
		ChannelId: request.ChannelId,
		RootId:    request.RootID,
		CreateAt:  time,
		UpdateAt:  time,
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
		api.LogWarn(message, "error", err)
	}
}
