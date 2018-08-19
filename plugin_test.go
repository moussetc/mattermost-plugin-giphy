package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
)

func initMockAPI() *GiphyPlugin {

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

	p := GiphyPlugin{}
	p.SetAPI(api)

	return &p
}

func TestExecuteCommandToReturnCommandResponse(t *testing.T) {
	p := initMockAPI()

	assert.Nil(t, p.OnActivate())

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

func TestExecuteCommandToReturDisabledPluginError(t *testing.T) {

	p := initMockAPI()
	p.gifProvider = &giphyProvider{}

	response, err := p.ExecuteCommand(&plugin.Context{}, &model.CommandArgs{Command: "/gif cute doggo"})
	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.True(t, strings.Contains(err.Error(), "disabled"))

	assert.Nil(t, p.OnDeactivate())
}

func TestExecuteCommandToReturUnableToGetGIFError(t *testing.T) {

	p := initMockAPI()
	assert.Nil(t, p.OnActivate())

	errorMessage := "ARGHHHH"
	p.gifProvider = &mockGifProviderFail{errorMessage}

	response, err := p.ExecuteCommand(&plugin.Context{}, &model.CommandArgs{Command: "/gif cute doggo"})
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

func (m *mockGifProviderFail) getMultipleGifsURL(config *GiphyPluginConfiguration, request string) ([]string, error) {
	return nil, errors.New(m.errorMessage)
}

// mockGifProvider always provides the same fake GIF URL
type mockGifProvider struct {
	mockURL string
}

func (m *mockGifProvider) getGifURL(config *GiphyPluginConfiguration, request string) (string, error) {
	return m.mockURL, nil
}

func (m *mockGifProvider) getMultipleGifsURL(config *GiphyPluginConfiguration, request string) ([]string, error) {
	return []string{m.mockURL, m.mockURL, m.mockURL}, nil
}
