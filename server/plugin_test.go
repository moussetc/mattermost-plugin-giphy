package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
)

func generateMockPluginConfig() configuration {
	return configuration{
		Provider:        "giphy",
		Language:        "fr",
		Rating:          "",
		Rendition:       "fixed_height_small",
		RenditionTenor:  "tinygif",
		RenditionGfycat: "gif100Px",
		APIKey:          "defaultAPIKey",
	}
}

func mockLoadConfig(conf configuration) func(dest interface{}) error {
	return func(dest interface{}) error {
		*dest.(*configuration) = conf
		return nil
	}
}

type mockHTTPHandler struct{}

func (h *mockHTTPHandler) handleCancel(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	w.WriteHeader(http.StatusOK)
}
func (h *mockHTTPHandler) handleShuffle(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	w.WriteHeader(http.StatusOK)
}
func (h *mockHTTPHandler) handleSend(p *Plugin, w http.ResponseWriter, request *integrationRequest) {
	w.WriteHeader(http.StatusOK)
}

func initMockAPI() (api *plugintest.API, p *Plugin) {
	api = &plugintest.API{}

	pluginConfig := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(pluginConfig))
	p = &Plugin{}
	p.SetAPI(api)
	p.botId = "botId42"
	p.httpHandler = &mockHTTPHandler{}
	return api, p
}

func setMockHelpers(plugin *Plugin) {
	testHelpers := &plugintest.Helpers{}
	testHelpers.On("EnsureBot", mock.AnythingOfType("*model.Bot"), mock.AnythingOfType("plugin.EnsureBotOption")).Return("botId42", nil)
	plugin.SetHelpers(testHelpers)
}

func TestOnActivateWithBadConfig(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	config.APIKey = ""
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(config))
	p := Plugin{}
	p.SetAPI(api)

	assert.NotNil(t, p.OnActivate())
}

func TestOnActivateOK(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(config))
	api.On("RegisterCommand", mock.Anything).Return(nil)
	p := Plugin{}
	p.SetAPI(api)
	setMockHelpers(&p)

	assert.Nil(t, p.OnActivate())
}

func TestExecuteGifCommandToSendPost(t *testing.T) {
	_, p := initMockAPI()

	url := "http://fakeURL"
	p.gifProvider = &mockGifProvider{url}

	command := model.CommandArgs{
		Command: "/gif cute doggo",
		UserId:  "userid",
	}

	response, err := p.ExecuteCommand(&plugin.Context{}, &command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, strings.Contains(response.Text, url))
	assert.Equal(t, "in_channel", response.ResponseType)
}

func TestExecuteShuffleCommandToReturnCommandResponse(t *testing.T) {
	api, p := initMockAPI()
	url := "http://fakeURL"
	p.gifProvider = &mockGifProvider{url}

	command := model.CommandArgs{
		Command: "/gifs cute doggo",
		UserId:  "userid",
	}
	api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil, nil)
	response, err := p.ExecuteCommand(&plugin.Context{}, &command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
}

func TestExecuteCommandToReturUnableToGetGIFError(t *testing.T) {
	_, p := initMockAPI()

	errorMessage := "ARGHHHH"
	p.gifProvider = &mockGifProviderFail{errorMessage}

	response, err := p.ExecuteCommand(&plugin.Context{}, &model.CommandArgs{Command: "/gif cute doggo"})
	assert.NotNil(t, err)
	assert.Empty(t, response)
	assert.True(t, strings.Contains(err.DetailedError, errorMessage))
}

func TestExecuteUnkownCommand(t *testing.T) {
	_, p := initMockAPI()

	command := model.CommandArgs{
		Command: "/worm cute doggo",
		UserId:  "userid",
	}

	response, err := p.ExecuteCommand(&plugin.Context{}, &command)
	assert.NotNil(t, err)
	assert.Nil(t, response)
}

func TestServeHTTP(t *testing.T) {
	p := setupMockPluginWithAuthent()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", URLShuffle, nil)
	r.Header.Add("Mattermost-User-Id", testUserId)
	p.ServeHTTP(nil, w, r)
	result := w.Result()
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.StatusCode)
}

// mockGifProviderFail always fail to provide a GIF URL
type mockGifProviderFail struct {
	errorMessage string
}

func (m *mockGifProviderFail) getGifURL(config *configuration, request string, cursor *string) (string, *model.AppError) {
	return "", appError(m.errorMessage, errors.New(m.errorMessage))
}

// mockGifProvider always provides the same fake GIF URL
type mockGifProvider struct {
	mockURL string
}

func (m *mockGifProvider) getGifURL(config *configuration, request string, cursor *string) (string, *model.AppError) {
	return m.mockURL, nil
}
