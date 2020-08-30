package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
)

const (
	testChannelId = "gifs-channel"
	testUserId    = "gif-user"
	testPostId    = "skfqsldjhfkljhf"
	testKeywords  = "kitty"
	testGifURL    = "https://gif.fr/gif/42"
	testCursor    = "43abc"
	testRootId    = "4242abc"
)

var testPostActionIntegrationRequest = model.PostActionIntegrationRequest{
	ChannelId: testChannelId,
	UserId:    testUserId,
	PostId:    testPostId,
	Context: map[string]interface{}{
		contextGifURL:   testGifURL,
		contextKeywords: testKeywords,
		contextCursor:   testCursor,
		contextRootId:   testRootId,
	},
}

func generateTestIntegrationRequest() *integrationRequest {
	return &integrationRequest{
		testGifURL, testKeywords, testCursor, testRootId, testPostActionIntegrationRequest,
	}
}

func mockPostActionIntegratioRequestFromJSON(body io.Reader) *model.PostActionIntegrationRequest {
	return &testPostActionIntegrationRequest
}

func generateMockAPIForHandlers() *plugintest.API {
	api := &plugintest.API{}
	api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil)
	api.On("LogWarn", mock.AnythingOfType("string"), nil).Return(nil)
	api.On("LogWarn", mock.AnythingOfType("string"), mock.AnythingOfType("*model.AppError")).Return(nil)
	return api
}

func setupMockPluginWithAuthent() *Plugin {
	api := &plugintest.API{}
	api.On("HasPermissionToChannel", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*model.Permission")).Return(true)
	p := &Plugin{}
	p.SetAPI(api)
	p.httpHandler = &mockHTTPHandler{}

	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON

	return p
}

func TestHandleHTTPRequestOK(t *testing.T) {
	p := setupMockPluginWithAuthent()

	goodURLs := [3]string{URLCancel, URLShuffle, URLSend}
	for _, URL := range goodURLs {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", URL, nil)
		r.Header.Add("Mattermost-User-Id", testUserId)

		p.handleHTTPRequest(w, r)

		result := w.Result()
		assert.NotNil(t, result)
		assert.Equal(t, 200, result.StatusCode)
	}
}

func TestHandleHTTPRequestKOBadMethod(t *testing.T) {
	p := setupMockPluginWithAuthent()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", URLSend, nil)
	r.Header.Add("Mattermost-User-Id", testUserId)

	p.handleHTTPRequest(w, r)

	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 405, result.StatusCode)
}

func TestHandleHTTPRequestKOBadURL(t *testing.T) {
	p := setupMockPluginWithAuthent()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/unexistingURL", nil)
	r.Header.Add("Mattermost-User-Id", testUserId)

	p.handleHTTPRequest(w, r)

	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 404, result.StatusCode)
}

func TestHandleHTTPRequestKOWithoutUser(t *testing.T) {
	p := setupMockPluginWithAuthent()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLSend, nil)

	p.handleHTTPRequest(w, r)

	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 401, result.StatusCode)
}

func TestHandleHTTPRequestKOBadUser(t *testing.T) {
	p := setupMockPluginWithAuthent()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLSend, nil)
	r.Header.Add("Mattermost-User-Id", "differentUserId")

	p.handleHTTPRequest(w, r)

	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 400, result.StatusCode)
}

func TestHandleHTTPRequestKOUserNotAllowedToPostHere(t *testing.T) {
	api := &plugintest.API{}
	api.On("HasPermissionToChannel",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("*model.Permission"),
	).Return(false)
	p := &Plugin{}
	p.SetAPI(api)
	p.httpHandler = &mockHTTPHandler{}

	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLSend, nil)
	r.Header.Add("Mattermost-User-Id", testUserId)

	p.handleHTTPRequest(w, r)

	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 403, result.StatusCode)
}

func TestParseRequestOK(t *testing.T) {
	api := generateMockAPIForHandlers()
	postActionIntegrationRequestFromJson = mockPostActionIntegratioRequestFromJSON

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLSend, nil)
	request, ok := parseRequest(api, w, r)
	assert.NotNil(t, request)
	assert.Equal(t, request.ChannelId, testChannelId)
	assert.Equal(t, request.UserId, testUserId)
	assert.Equal(t, request.GifURL, testGifURL)
	assert.Equal(t, request.Keywords, testKeywords)
	assert.Equal(t, request.Cursor, testCursor)
	assert.Equal(t, request.RootId, testRootId)
	assert.True(t, ok)
}

func TestParseRequestKOBadRequest(t *testing.T) {
	api := generateMockAPIForHandlers()
	postActionIntegrationRequestFromJson = func(body io.Reader) *model.PostActionIntegrationRequest { return nil }

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLSend, nil)
	request, ok := parseRequest(api, w, r)
	assert.Nil(t, request)
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
				ChannelId: testChannelId,
				UserId:    testUserId,
				PostId:    testPostId,
				Context:   context,
			}
		}

		notifyHandlerError = func(api plugin.API, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
			assert.Contains(t, message, contextElements[i])
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", URLSend, nil)
		request, ok := parseRequest(api, w, r)
		assert.Nil(t, request)
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
	api.On("DeleteEphemeralPost",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).Return(nil)
	p := Plugin{}
	p.SetAPI(api)
	h := &defaultHTTPHandler{}
	w := httptest.NewRecorder()
	h.handleCancel(&p, w, generateTestIntegrationRequest())
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	api.AssertCalled(t,
		"DeleteEphemeralPost",
		mock.MatchedBy(func(s string) bool { return s == testUserId }),
		mock.MatchedBy(func(postId string) bool { return postId == testPostId }))
}

func TestHandleShuffleOK(t *testing.T) {
	api := &plugintest.API{}
	api.On("UpdateEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil)
	p := Plugin{}
	p.SetAPI(api)
	p.gifProvider = &mockGifProvider{"fakeURL"}
	h := &defaultHTTPHandler{}
	w := httptest.NewRecorder()
	h.handleShuffle(&p, w, generateTestIntegrationRequest())
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	api.AssertCalled(t, "UpdateEphemeralPost",
		mock.MatchedBy(func(s string) bool { return s == testUserId }),
		mock.MatchedBy(func(post *model.Post) bool {
			return post.Id == testPostId &&
				strings.Contains(post.Message, "fakeURL") &&
				post.UserId == p.botId &&
				post.ChannelId == testChannelId &&
				post.RootId == testRootId
		}))
}

func TestHandleShuffleKOProviderError(t *testing.T) {
	api := &plugintest.API{}
	p := Plugin{}
	p.SetAPI(api)
	p.gifProvider = &mockGifProviderFail{"fakeURL"}
	h := &defaultHTTPHandler{}

	notifyHandlerError = func(api plugin.API, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
		assert.Contains(t, message, "Gif")
	}

	w := httptest.NewRecorder()
	h.handleShuffle(&p, w, generateTestIntegrationRequest())
	assert.Equal(t, w.Result().StatusCode, http.StatusServiceUnavailable)
	api.AssertNumberOfCalls(t, "UpdateEphemeralPost", 0)
}

func TestHandleSendOK(t *testing.T) {
	api := &plugintest.API{}
	api.On("DeleteEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, nil)
	p := Plugin{}
	p.SetAPI(api)
	p.gifProvider = &mockGifProvider{"fakeURL"}
	h := &defaultHTTPHandler{}
	w := httptest.NewRecorder()
	h.handleSend(&p, w, generateTestIntegrationRequest())
	assert.Equal(t, w.Result().StatusCode, http.StatusOK)
	api.AssertCalled(t,
		"DeleteEphemeralPost",
		mock.MatchedBy(func(s string) bool { return s == testUserId }),
		mock.MatchedBy(func(postId string) bool { return postId == testPostId }))
	api.AssertCalled(t,
		"CreatePost",
		mock.MatchedBy(func(p *model.Post) bool {
			return strings.Contains(p.Message, testGifURL) &&
				p.UserId == testUserId &&
				p.ChannelId == testChannelId &&
				p.RootId == testRootId
		}),
	)

}

func TestHandleSendKOCreatePostError(t *testing.T) {
	api := &plugintest.API{}
	api.On("DeleteEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, model.NewAppError("test", "id42", nil, "errorMessage", 42))
	notifyHandlerError = func(api plugin.API, message string, err *model.AppError, request *model.PostActionIntegrationRequest) {
		assert.Contains(t, message, "create")
	}

	p := Plugin{}
	p.SetAPI(api)
	p.gifProvider = &mockGifProvider{"fakeURL"}
	h := &defaultHTTPHandler{}
	w := httptest.NewRecorder()
	h.handleSend(&p, w, generateTestIntegrationRequest())
	assert.Equal(t, w.Result().StatusCode, http.StatusInternalServerError)
}

func TestDefaultNotifyHandlerErrorOK(t *testing.T) {
	api := &plugintest.API{}
	api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil)
	api.On("LogWarn", mock.Anything, mock.Anything).Return(nil)
	message := "oops"
	err := test.MockErrorGenerator().FromError(message, errors.New("strange failure"))
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
