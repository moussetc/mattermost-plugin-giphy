package main

import (
	"errors"
	"testing"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

func generateMocksForConfigurationTesting(displayMode string) *Plugin {
	api := &plugintest.API{}
	pluginConfig := generateMockPluginConfig()
	pluginConfig.DisplayMode = displayMode
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(pluginConfig))
	p := Plugin{}
	p.errorGenerator = test.MockErrorGenerator()
	p.SetAPI(api)
	return &p
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
	p := generateMocksForConfigurationTesting("")
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
