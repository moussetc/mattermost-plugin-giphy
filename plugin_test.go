package main

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
)

func TestGetGifUrl(t *testing.T) {
	p := &GiphyPlugin{}
	p.configuration.Store(&GiphyPluginConfiguration{Language: "fr", Rating: "", Rendition: "fixed_height_small"})
	text, err := p.getGifURL("cat")
	if err != nil {
		t.Errorf("Error while retrieving gif %v", err)
	}
	if text == "" {
		t.Errorf("Text is empty, it should be a GIF URL")
	}
}

func TestPlugin(t *testing.T) {

	configuration := GiphyPluginConfiguration{
		Language:  "fr",
		Rating:    "",
		Rendition: "fixed_height_small",
	}

	api := &plugintest.API{}
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.GiphyPluginConfiguration")).Return(func(dest interface{}) error {
		*dest.(*GiphyPluginConfiguration) = configuration
		return nil
	})
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)

	p := GiphyPlugin{}
	assert.Nil(t, p.OnActivate(api))

	command := &model.CommandArgs{
		Command: "/gif cute doggo",
	}
	response, err := p.ExecuteCommand(command)
	assert.Nil(t, err)
	assert.NotNil(t, response)

	assert.Nil(t, p.OnDeactivate())
}
