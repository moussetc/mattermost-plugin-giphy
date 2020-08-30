package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/moussetc/mattermost-plugin-giphy/server/internal/provider"

	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

func generateMocksForConfigurationTesting(displayMode, gifProvider string) *Plugin {
	api := &plugintest.API{}
	pluginConfig := generateMockPluginConfig()
	pluginConfig.Provider = gifProvider
	pluginConfig.DisplayMode = displayMode
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(pluginConfig))
	p := Plugin{}
	p.SetAPI(api)
	setMockHelpers(&p)
	return &p
}

func TestOnConfigurationChangeLoadFail(t *testing.T) {
	api := &plugintest.API{}
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(errors.New("Failed config load"))
	p := Plugin{}
	p.SetAPI(api)

	err := p.OnConfigurationChange()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed config load")
}

func TestOnConfigurationChangeEmptyDisplayMode(t *testing.T) {
	p := generateMocksForConfigurationTesting("", "giphy")
	err := p.OnConfigurationChange()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The Display Mode must be configured")
}

func TestOnConfigurationChangeEmptyProvider(t *testing.T) {
	p := generateMocksForConfigurationTesting("embedded", "")
	err := p.OnConfigurationChange()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The GIF provider must be configured")
}

func TestOnConfigurationChangeGiphyProvider(t *testing.T) {
	p := generateMocksForConfigurationTesting("embedded", "giphy")

	err := p.OnConfigurationChange()
	assert.Nil(t, err)
	assert.NotNil(t, p.gifProvider)
	assert.Equal(t, reflect.TypeOf(&provider.Giphy{}).String(), reflect.TypeOf(p.gifProvider).String())
}

func TestOnConfigurationChangeGfycatProvider(t *testing.T) {
	p := generateMocksForConfigurationTesting("embedded", "gfycat")

	err := p.OnConfigurationChange()
	assert.Nil(t, err)
	assert.NotNil(t, p.gifProvider)
	assert.Equal(t, reflect.TypeOf(&provider.Gfycat{}).String(), reflect.TypeOf(p.gifProvider).String())
}

func TestOnConfigurationChangeTenorProvider(t *testing.T) {
	p := generateMocksForConfigurationTesting("embedded", "tenor")

	err := p.OnConfigurationChange()
	assert.Nil(t, err)
	assert.NotNil(t, p.gifProvider)
	assert.Equal(t, reflect.TypeOf(&provider.Tenor{}).String(), reflect.TypeOf(p.gifProvider).String())
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
