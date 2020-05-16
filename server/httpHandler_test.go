package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
)

func mockPostActionIntegratioRequestFromJSON(body io.Reader) *model.PostActionIntegrationRequest {
	return &model.PostActionIntegrationRequest{
		ChannelId: "channelID",
		UserId:    "userID",
		PostId:    "postID",
		Context: map[string]interface{}{
			contextGifURL:   "gifUrl",
			contextKeywords: "keywords",
			contextCursor:   "cursor",
			contextRootId:   "rootID",
		},
	}
}

func generateMockAPIForHandlers() *plugintest.API {
	api := &plugintest.API{}
	api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil)
	api.On("LogWarn", mock.AnythingOfType("string"), nil).Return(nil)
	api.On("LogWarn", mock.AnythingOfType("string"), mock.AnythingOfType("*model.AppError")).Return(nil)
	return api
}

func TestHandleHTTPRequest(t *testing.T) {
	p := &Plugin{}
	p.httpHandler = &mockHTTPHandler{}

	goodURLs := [3]string{URLCancel, URLShuffle, URLSend}
	for _, URL := range goodURLs {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", URL, nil)
		p.handleHTTPRequest(w, r)
		result := w.Result()
		assert.NotNil(t, result)
		assert.Equal(t, 200, result.StatusCode)
	}
}

func TestHandleHTTPRequestBadMethod(t *testing.T) {
	p := &Plugin{}
	p.httpHandler = &mockHTTPHandler{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", URLSend, nil)

	p.handleHTTPRequest(w, r)

	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 405, result.StatusCode)
}

func TestHandleHTTPRequestBadURL(t *testing.T) {
	p := &Plugin{}
	p.httpHandler = &mockHTTPHandler{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/unexistingURL", nil)

	p.handleHTTPRequest(w, r)

	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 404, result.StatusCode)
}


func TestParseRequestOK(t *testing.T) {
	api := generateMockAPIForHandlers()
	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLSend, nil)
	request, url, keywords, cursor, rootId, ok := parseRequest(api, w, r)
	assert.NotNil(t, request)
	assert.Equal(t, request.ChannelId, "channelID")
	assert.Equal(t, request.UserId, "userID")
	assert.Equal(t, url, "gifUrl")
	assert.Equal(t, keywords, "keywords")
	assert.Equal(t, cursor, "cursor")
	assert.Equal(t, rootId, "rootID")
	assert.True(t, ok)
}

func TestParseRequestKOBadRequest(t *testing.T) {
	api := generateMockAPIForHandlers()
	postActionIntegrationRequestFromJson = func(body io.Reader) *model.PostActionIntegrationRequest { return nil }

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLSend, nil)
	request, url, keywords, cursor, rootId, ok := parseRequest(api, w, r)
	assert.Nil(t, request)
	assert.Empty(t, url, "gifUrl")
	assert.Empty(t, keywords, "keywords")
	assert.Empty(t, cursor, "cursor")
	assert.Empty(t, rootId, "rootID")
	assert.False(t, ok)
	api.AssertCalled(t, "LogWarn", mock.MatchedBy(func(s string) bool { return strings.Contains(s, "parse") }), nil)
}

func TestParseRequestKOMissingContext(t *testing.T) {
	contextElements := [3]string{contextGifURL, contextKeywords, contextCursor}
	api := generateMockAPIForHandlers()
	for i := 0; i < len(contextElements); i++ {
		context := map[string]interface{}{}
		for j := 0; j < len(contextElements); j++ {
			if j != i {
				context[contextElements[j]] = contextElements[j]
			}
		}

		postActionIntegrationRequestFromJson = func(body io.Reader) *model.PostActionIntegrationRequest {
			return &model.PostActionIntegrationRequest{
				ChannelId: "channelID",
				UserId:    "userID",
				Context:   context,
			}
		}

		notifyHandlerError = func(api plugin.API, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
			assert.Contains(t, message, contextElements[i])
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", URLSend, nil)
		request, url, keywords, cursor, rootId, ok := parseRequest(api, w, r)
		assert.Nil(t, request)
		assert.Empty(t, url, "gifUrl")
		assert.Empty(t, keywords, "keywords")
		assert.Empty(t, cursor, "cursor")
		assert.Empty(t, rootId, "rootID")
		assert.False(t, ok)
	}
}

func TestWriteResponseOKStatus(t *testing.T) {
	w := httptest.NewRecorder()
	writeResponse(http.StatusOK, w)
	result := w.Result()
	assert.Equal(t, result.StatusCode, http.StatusOK)
	bodyBytes, _ := ioutil.ReadAll(result.Body)
	assert.Contains(t, string(bodyBytes), "update")
}

func TestWriteResponseOtherStatus(t *testing.T) {
	w := httptest.NewRecorder()
	writeResponse(http.StatusTeapot, w)
	result := w.Result()
	assert.Equal(t, result.StatusCode, http.StatusTeapot)
	bodyBytes, _ := ioutil.ReadAll(result.Body)
	assert.NotContains(t, string(bodyBytes), "update")
}

func TestHandleCancel(t *testing.T) {
	api := &plugintest.API{}
	api.On("DeleteEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	p := Plugin{}
	p.SetAPI(api)
	h := &defaultHTTPHandler{}
	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLCancel, nil)
	h.handleCancel(&p, w, r)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	api.AssertCalled(t, "DeleteEphemeralPost", mock.MatchedBy(func(s string) bool { return s == "userID" }), mock.MatchedBy(func(postId string) bool { return postId == "postID" }))
}

func TestHandleShuffleOK(t *testing.T) {
	api := &plugintest.API{}
	api.On("UpdateEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil)
	p := Plugin{}
	p.SetAPI(api)
	p.gifProvider = &mockGifProvider{"fakeURL"}
	h := &defaultHTTPHandler{}
	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLShuffle, nil)
	h.handleShuffle(&p, w, r)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	api.AssertCalled(t, "UpdateEphemeralPost", mock.MatchedBy(func(s string) bool { return s == "userID" }), mock.MatchedBy(func(p *model.Post) bool {
		return p.Id == "postID" &&
			strings.Contains(p.Message, "fakeURL") &&
			p.UserId == "userID" &&
			p.ChannelId == "channelID" &&
			p.RootId == "rootID"
	}))
}

func TestHandleShuffleKOProviderError(t *testing.T) {
	api := &plugintest.API{}
	p := Plugin{}
	p.SetAPI(api)
	p.gifProvider = &mockGifProviderFail{"fakeURL"}
	h := &defaultHTTPHandler{}
	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON

	notifyHandlerError = func(api plugin.API, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
		assert.Contains(t, message, "Gif")
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLShuffle, nil)
	h.handleShuffle(&p, w, r)
	assert.Equal(t, w.Result().StatusCode, http.StatusServiceUnavailable)
	api.AssertNumberOfCalls(t, "UpdateEphemeralPost", 0)
}

func TestHandlePostOK(t *testing.T) {
	api := &plugintest.API{}
	api.On("DeleteEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, nil)
	p := Plugin{}
	p.SetAPI(api)
	h := &defaultHTTPHandler{}
	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLCancel, nil)
	h.handlePost(&p, w, r)
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	api.AssertCalled(t, "DeleteEphemeralPost", mock.MatchedBy(func(s string) bool { return s == "userID" }), mock.MatchedBy(func(postId string) bool { return postId == "postID" }))
	api.AssertCalled(t, "CreatePost", mock.MatchedBy(func(p *model.Post) bool {
		return strings.Contains(p.Message, "gifUrl") && p.UserId == "userID" && p.ChannelId == "channelID" && p.RootId == "rootID"
	}))
}

func TestHandlePostKOCreatePostError(t *testing.T) {
	api := &plugintest.API{}
	api.On("DeleteEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, appError("errorMessage", nil))
	notifyHandlerError = func(api plugin.API, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
		assert.Contains(t, message, "create")
	}

	p := Plugin{}
	p.SetAPI(api)
	h := &defaultHTTPHandler{}
	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLCancel, nil)
	h.handlePost(&p, w, r)
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
}

func TestDefaultNotifyHandlerErrorOK(t *testing.T) {
	api := &plugintest.API{}
	api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil)
	api.On("LogWarn", mock.Anything, mock.Anything).Return(nil)
	message := "oops"
	err := appError(message, errors.New("strange failure"))
	channelID := "42"
	userID := "jane"
	request := &model.PostActionIntegrationRequest{
		UserId:    userID,
		ChannelId: channelID,
	}
	userIDCheck := func(uID string) bool { return uID == userID }
	postCheck := func(post *model.Post) bool {
		return post != nil &&
			post.ChannelId == channelID &&
			strings.Contains(post.Message, message) &&
			(err == nil || strings.Contains(post.Message, err.Message))
	}
	// With error
	defaultNotifyHandlerError(api, message, err, request)
	api.AssertNumberOfCalls(t, "LogWarn", 1)
	api.AssertCalled(t, "SendEphemeralPost",
		mock.MatchedBy(userIDCheck),
		mock.MatchedBy(postCheck))

	// Without error
	err = nil
	defaultNotifyHandlerError(api, message, err, request)
	api.AssertNumberOfCalls(t, "LogWarn", 2)
	api.AssertCalled(t, "SendEphemeralPost",
		mock.MatchedBy(userIDCheck),
		mock.MatchedBy(postCheck))
}
