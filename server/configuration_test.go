package main

import (
	"errors"
	"testing"

	manifest "github.com/moussetc/mattermost-plugin-giphy"
	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

func generateMocksForConfigurationTesting(pluginConfig *pluginConf.Configuration) *Plugin {
	api := &plugintest.API{}
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(*pluginConfig))
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)

	siteURL := "https://test.com"
	serverConfig := &model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &siteURL,
		},
	}

	api.On("GetConfig").Return(serverConfig)
	p := Plugin{}
	p.errorGenerator = test.MockErrorGenerator()
	p.SetAPI(api)
	p.setConfiguration(pluginConfig)
	return &p
}

func TestOnConfigurationChangeOK(t *testing.T) {
	configuration := generateMockPluginConfig()
	configuration.DisplayMode = pluginConf.DisplayModeEmbedded
	configuration.Provider = "giphy"
	p := generateMocksForConfigurationTesting(&configuration)

	err := p.OnConfigurationChange()

	assert.Nil(t, err)
	assert.Equal(t, "https://test.com/plugins/"+manifest.Manifest.Id, p.rootURL)
}

func TestOnConfigurationChangeLoadFail(t *testing.T) {
	api := &plugintest.API{}
	mockErr := errors.New("failed config load")
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockErr)
	p := Plugin{}
	p.SetAPI(api)

	err := p.OnConfigurationChange()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), mockErr.Error())
}

func TestOnConfigurationChangeEmptyDisplayMode(t *testing.T) {
	configuration := generateMockPluginConfig()
	configuration.DisplayMode = ""
	p := generateMocksForConfigurationTesting(&configuration)

	err := p.OnConfigurationChange()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "the Display Mode must be configured")
}

func TestOnConfigurationChangeGifProviderError(t *testing.T) {
	api := &plugintest.API{}
	pluginConfig := generateMockPluginConfig()
	pluginConfig.DisplayMode = pluginConf.DisplayModeEmbedded
	pluginConfig.APIKey = ""
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(pluginConfig))

	p := Plugin{errorGenerator: test.MockErrorGenerator()}
	p.SetAPI(api)
	err := p.OnConfigurationChange()
	assert.NotNil(t, err)
}

func TestGetSetConfiguration(t *testing.T) {
	p := Plugin{}

	initialConfig := p.getConfiguration()
	assert.NotNil(t, initialConfig)

	initialConfig.APIKey = "COUCOU"
	p.setConfiguration(initialConfig)

	modifiedConfig := p.getConfiguration()
	assert.Equal(t, initialConfig.APIKey, modifiedConfig.APIKey)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	p.setConfiguration(modifiedConfig)
}
