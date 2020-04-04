package main

import (
	"errors"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestOnConfigurationChangeLoadFail(t *testing.T) {
	api := &plugintest.API{}
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(errors.New("Failed config load"))
	p := Plugin{}
	p.SetAPI(api)

	err := p.OnConfigurationChange()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed config load")
}

func TestOnConfigurationChangeEmptyProvider(t *testing.T) {
	api := &plugintest.API{}
	pluginConfig := generateMockPluginConfig()
	pluginConfig.Provider = ""
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(pluginConfig))
	p := Plugin{}
	p.SetAPI(api)

	err := p.OnConfigurationChange()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "The GIF provider must be configured")
}

func TestOnConfigurationChangeGiphyProvider(t *testing.T) {
	api := &plugintest.API{}
	pluginConfig := generateMockPluginConfig()
	pluginConfig.Provider = "giphy"
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(pluginConfig))
	p := Plugin{}
	p.SetAPI(api)

	err := p.OnConfigurationChange()
	assert.Nil(t, err)
	assert.NotNil(t, p.gifProvider)
	assert.Equal(t, reflect.TypeOf(&giphyProvider{}).String(), reflect.TypeOf(p.gifProvider).String())
}

func TestOnConfigurationChangeGfycatProvider(t *testing.T) {
	api := &plugintest.API{}
	pluginConfig := generateMockPluginConfig()
	pluginConfig.Provider = "gfycat"
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(pluginConfig))
	p := Plugin{}
	p.SetAPI(api)

	err := p.OnConfigurationChange()
	assert.Nil(t, err)
	assert.NotNil(t, p.gifProvider)
	assert.Equal(t, reflect.TypeOf(&gfyCatProvider{}).String(), reflect.TypeOf(p.gifProvider).String())
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
