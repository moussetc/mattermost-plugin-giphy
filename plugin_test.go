package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
)

func initMockAPI() *plugintest.API {

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

	return api
}

func TestExecuteCommandToReturnCommandResponse(t *testing.T) {
	api := initMockAPI()

	p := GiphyPlugin{}
	assert.Nil(t, p.OnActivate(api))

	url := "http://fakeURL"
	p.gifProvider = &mockGifProvider{url}

	command := model.CommandArgs{
		Command: "/gif cute doggo",
		UserId:  "userid",
	}

	response, err := p.ExecuteCommand(&command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, strings.Contains(response.Text, url))
	assert.Equal(t, "in_channel", response.ResponseType)
}

func TestExecuteCommandToReturDisabledPluginError(t *testing.T) {
	api := initMockAPI()

	p := GiphyPlugin{}
	p.api = api
	p.gifProvider = &giphyProvider{}

	response, err := p.ExecuteCommand(&model.CommandArgs{Command: "/gif cute doggo"})
	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.True(t, strings.Contains(err.Error(), "disabled"))

	assert.Nil(t, p.OnDeactivate())
}

func TestExecuteCommandToReturUnableToGetGIFError(t *testing.T) {
	api := initMockAPI()

	p := GiphyPlugin{}
	assert.Nil(t, p.OnActivate(api))

	errorMessage := "ARGHHHH"
	p.gifProvider = &mockGifProviderFail{errorMessage}

	response, err := p.ExecuteCommand(&model.CommandArgs{Command: "/gif cute doggo"})
	assert.NotNil(t, err)
	assert.Empty(t, response)
	assert.True(t, strings.Contains(err.DetailedError, errorMessage))
}

// mockGifProviderFail always fail to provide a GIF URL
type mockGifProviderFail struct {
	errorMessage string
}

func (m *mockGifProviderFail) getGifURL(config *GiphyPluginConfiguration, request string) (string, error) {
	return "", errors.New(m.errorMessage)
}

// mockGifProvider always provides the same fake GIF URL
type mockGifProvider struct {
	mockURL string
}

func (m *mockGifProvider) getGifURL(config *GiphyPluginConfiguration, request string) (string, error) {
	return m.mockURL, nil
}
