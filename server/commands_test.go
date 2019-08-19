package main

import (
	"errors"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestRegisterCommandsKORegisterGifCommand(t *testing.T) {
	api := &plugintest.API{}
	api.On("GetConfig").Return(generateMockMattermostConfig())
	config := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(config))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == "gif" })).Return(errors.New("Fail mock register command"))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == "gifs" })).Return(nil)
	p := Plugin{}
	p.SetAPI(api)

	assert.NotNil(t, p.RegisterCommands())
}

func TestRegisterCommandsKORegisterGifsCommand(t *testing.T) {
	api := &plugintest.API{}
	api.On("GetConfig").Return(generateMockMattermostConfig())
	config := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(mockLoadConfig(config))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == "gif" })).Return(nil)
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == "gifs" })).Return(errors.New("Fail mock register command"))
	p := Plugin{}
	p.SetAPI(api)

	assert.NotNil(t, p.RegisterCommands())
}

func TestExecuteCommandGifOK(t *testing.T) {
	p := initMockAPI()
	keywords := "coucou"
	p.gifProvider = &mockGifProvider{}
	response, err := p.executeCommandGif(keywords)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, strings.Contains(response.Text, keywords))
}

func TestExecuteCommandGifUnableToGetGIFError(t *testing.T) {
	p := initMockAPI()

	errorMessage := "ARGHHHH"
	p.gifProvider = &mockGifProviderFail{errorMessage}

	response, err := p.executeCommandGif("mayhem")
	assert.NotNil(t, err)
	assert.Empty(t, response)
	assert.Contains(t, err.DetailedError, errorMessage)
}

func TestExecuteCommandGifShuffleOK(t *testing.T) {
	p := Plugin{}
	p.gifProvider = &mockGifProvider{}
	command := "/gifs hello"
	response, err := p.executeCommandGifShuffle(command, nil)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, model.COMMAND_RESPONSE_TYPE_EPHEMERAL, response.ResponseType)
	assert.Contains(t, response.Text, "hello")
	assert.NotNil(t, response.Attachments) // attachments content is tested in the generate function test
}

func TestExecuteCommandGifShuffleKOProviderError(t *testing.T) {
	p := Plugin{}
	p.gifProvider = &mockGifProviderFail{"mockError"}
	command := "/gifs hello"
	response, err := p.executeCommandGifShuffle(command, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "mockError")
	assert.Nil(t, response)
}
