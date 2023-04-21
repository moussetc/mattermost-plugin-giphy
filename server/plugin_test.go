package main

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	manifest "github.com/moussetc/mattermost-plugin-giphy"
	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	pluginapi "github.com/moussetc/mattermost-plugin-giphy/server/internal/pluginapi"
	mock_pluginapi "github.com/moussetc/mattermost-plugin-giphy/server/internal/pluginapi/mock_pluginapi"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest/mock"
)

func generateMockPluginConfig() pluginConf.Configuration {
	return pluginConf.Configuration{
		DisplayMode:                  pluginConf.DisplayModeEmbedded,
		Provider:                     "giphy",
		Language:                     "fr",
		Rating:                       "none",
		Rendition:                    "fixed_height_small",
		RenditionTenor:               "tinygif",
		RenditionGfycat:              "gif100Px",
		APIKey:                       "defaultAPIKey",
		CommandTriggerGif:            triggerGif,
		CommandTriggerGifWithPreview: triggerGifs,
	}
}

func mockLoadConfig(conf pluginConf.Configuration) func(dest interface{}) error {
	return func(dest interface{}) error {
		*dest.(*pluginConf.Configuration) = conf
		return nil
	}
}

type mockHTTPHandler struct{}

func (h *mockHTTPHandler) handleCancel(_ *Plugin, w http.ResponseWriter, _ *integrationRequest) {
	w.WriteHeader(http.StatusOK)
}
func (h *mockHTTPHandler) handleShuffle(_ *Plugin, w http.ResponseWriter, _ *integrationRequest) {
	w.WriteHeader(http.StatusOK)
}
func (h *mockHTTPHandler) handlePrevious(_ *Plugin, w http.ResponseWriter, _ *integrationRequest) {
	w.WriteHeader(http.StatusOK)
}
func (h *mockHTTPHandler) handleSend(_ *Plugin, w http.ResponseWriter, _ *integrationRequest) {
	w.WriteHeader(http.StatusOK)
}

func initMockAPI() (api *plugintest.API, p *Plugin) {
	api = &plugintest.API{}

	pluginConfig := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(pluginConfig))
	p = &Plugin{}
	p.configuration = &pluginConfig
	p.SetAPI(api)
	p.botID = "botId42"
	p.httpHandler = &mockHTTPHandler{}
	p.errorGenerator = test.MockErrorGenerator()
	return api, p
}

func TestGeneratedManifestShouldBeValid(t *testing.T) {
	assert.Nil(t, manifest.Manifest.IsValid())
}

func TestOnActivateWithBadConfig(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	config.DisplayMode = ""
	config.APIKey = ""
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(config))
	siteURL := "https://test.com"
	serverConfig := &model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &siteURL,
		},
	}
	api.On("GetConfig").Return(serverConfig)
	p := Plugin{}
	p.SetAPI(api)

	assert.NotNil(t, p.OnActivate())
}

func TestOnActivateWithoutSiteURL(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	config.APIKey = ""
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(config))
	serverConfig := &model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: nil,
		},
	}
	api.On("GetConfig").Return(serverConfig)
	p := Plugin{}
	p.SetAPI(api)

	assert.NotNil(t, p.OnActivate())
}

func TestOnActivateOK(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	api.On("GetServerVersion").Return("42.0.0")
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)
	api.On("EnsureBotUser", mock.Anything).Return(model.NewId(), nil)
	api.On("GetServerVersion").Return("6.3.0")
	siteURL := "https://test.com"
	serverConfig := &model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &siteURL,
		},
	}
	api.On("GetConfig").Return(serverConfig)
	p := Plugin{}
	p.configuration = &config
	p.SetAPI(api)
	p.errorGenerator = test.MockErrorGenerator()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBot := mock_pluginapi.NewMockBotService(mockCtrl)
	mockBot.EXPECT().EnsureBot(gomock.Any(), gomock.Any())
	p.pluginClient = &pluginapi.Client{Bot: mockBot}

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

func TestExecuteCommandFailWhenCommandHandlerFails(t *testing.T) {
	api, p := initMockAPI()
	api.On("LogWarn", mock.AnythingOfType("string")).Return(nil)

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

// mockGifProviderFail always fail to provide a GIF URL
type mockGifProviderFail struct {
	errorMessage string
}

func (m *mockGifProviderFail) GetGifURL(_ string, _ *string, _ bool) ([]string, *model.AppError) {
	return []string{""}, (test.MockErrorGenerator()).FromError(m.errorMessage, errors.New(m.errorMessage))
}

func (m *mockGifProviderFail) GetAttributionMessage() string {
	return "test"
}

// mockGifProvider always provides the same fake GIF URL
type emptyGifProvider struct {
}

func (m *emptyGifProvider) GetGifURL(_ string, _ *string, _ bool) ([]string, *model.AppError) {
	return []string{}, nil
}

func (m *emptyGifProvider) GetAttributionMessage() string {
	return "test"
}

// mockGifProvider always provides the same fake GIF URL
type mockGifProvider struct {
	mockURL string
}

func newMockGifProvider() *mockGifProvider {
	return &mockGifProvider{"fakeURL"}
}

func (m *mockGifProvider) GetGifURL(_ string, _ *string, _ bool) ([]string, *model.AppError) {
	return []string{m.mockURL}, nil
}

func (m *mockGifProvider) GetAttributionMessage() string {
	return "test"
}
