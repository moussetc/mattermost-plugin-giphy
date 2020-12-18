package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
)

func TestRegisterCommandsKORegisterGifCommand(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(config))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGif })).Return(errors.New("Fail mock register command"))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGifWithPreview })).Return(nil)
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	p := Plugin{}
	p.configuration = &config
	p.SetAPI(api)

	assert.NotNil(t, p.RegisterCommands())
}

func TestRegisterCommandsKORegisterGifsCommand(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(config))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGif })).Return(nil)
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGifWithPreview })).Return(errors.New("Fail mock register command"))
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)
	p := Plugin{}
	p.configuration = &config
	p.SetAPI(api)

	assert.NotNil(t, p.RegisterCommands())
}

func TestExecuteCommandGifOK(t *testing.T) {
	_, p := initMockAPI()
	keywords := "coucou"
	caption := "hello"
	p.gifProvider = &mockGifProvider{}
	response, err := p.executeCommandGif(keywords, caption)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, strings.Contains(response.Text, keywords))
	assert.True(t, strings.Contains(response.Text, caption))
}

func TestExecuteCommandGifUnableToGetGIFError(t *testing.T) {
	_, p := initMockAPI()

	errorMessage := "ARGHHHH"
	p.gifProvider = &mockGifProviderFail{errorMessage}

	response, err := p.executeCommandGif("mayhem", "guy")
	assert.NotNil(t, err)
	assert.Empty(t, response)
	assert.Contains(t, err.DetailedError, errorMessage)
}

func TestExecuteCommandGifShuffleOK(t *testing.T) {
	api, p := initMockAPI()
	p.gifProvider = &mockGifProvider{}

	var recordCreationPost *model.Post
	api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil, nil).Run(func(args mock.Arguments) {
		recordCreationPost = args.Get(1).(*model.Post)
	})

	args := &model.CommandArgs{
		RootId:    "42",
		ChannelId: "43",
	}
	response, err := p.executeCommandGifWithPreview(testKeywords, testCaption, args)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", response.ResponseType)
	assert.NotNil(t, recordCreationPost)
	assert.True(t, strings.Contains(recordCreationPost.Message, testKeywords))
	assert.True(t, strings.Contains(recordCreationPost.Message, testCaption))
	assert.Equal(t, recordCreationPost.RootId, args.RootId)
	assert.Equal(t, recordCreationPost.ChannelId, args.ChannelId)
}

func TestExecuteCommandGifShuffleKOProviderError(t *testing.T) {
	p := Plugin{}
	p.gifProvider = &mockGifProviderFail{"mockError"}
	response, err := p.executeCommandGifWithPreview("hello", "", nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "mockError")
	assert.Nil(t, response)
}

func TestGenerateShufflePostAttachments(t *testing.T) {
	keywords := "rain"
	caption := "sad doctor"
	gifURL := "https://test.fr/rain-gif"
	cursor := "424242"
	rootId := "42"
	attachments := generateShufflePostAttachments(keywords, caption, gifURL, cursor, rootId)
	assert.NotNil(t, attachments)
	assert.Len(t, attachments, 1)
	attachment := attachments[0]
	assert.NotNil(t, attachment)
	actions := attachment.Actions
	assert.NotNil(t, actions)
	assert.Len(t, actions, 3)
	for i := 0; i < 3; i++ {
		assert.NotNil(t, actions[i].Integration)
		context := actions[i].Integration.Context
		assert.NotNil(t, context)
		assert.Equal(t, context[contextKeywords], keywords)
		assert.Equal(t, context[contextGifURL], gifURL)
		assert.Equal(t, context[contextCursor], cursor)
		assert.Equal(t, context[contextRootId], rootId)
	}
}

func TestParseCommandeLine(t *testing.T) {
	testCases := []struct {
		command          string
		expectedError    bool
		expectedKeywords string
		expectedCaption  string
	}{
		{command: "", expectedError: true, expectedKeywords: "", expectedCaption: ""},
		{command: "\"k1 k2 k3", expectedError: true, expectedKeywords: "", expectedCaption: ""},
		{command: "k1 k2 k3\"", expectedError: true, expectedKeywords: "", expectedCaption: ""},
		{command: "\"k1 k2 k3\" m1 m2 m3", expectedError: true, expectedKeywords: "", expectedCaption: ""},
		{command: "\"k1 k2 k3\" \"m1 m2 m3", expectedError: true, expectedKeywords: "", expectedCaption: ""},
		{command: "\"k1 k2 k3\" m1 m2 m3\"", expectedError: true, expectedKeywords: "", expectedCaption: ""},
		{command: "\"\" \"m1 m2 m3\"", expectedError: true, expectedKeywords: "", expectedCaption: ""},
		{command: "unique", expectedError: false, expectedKeywords: "unique", expectedCaption: ""},
		{command: "k1 k2", expectedError: false, expectedKeywords: "k1 k2", expectedCaption: ""},
		{command: "\"k1 k2 k3\"", expectedError: false, expectedKeywords: "k1 k2 k3", expectedCaption: ""},
		{command: "unique \"m1 m2 m3\"", expectedError: false, expectedKeywords: "unique", expectedCaption: "m1 m2 m3"},
		{command: "\"k1 k2 k3\" \"m1 m2 m3\"", expectedError: false, expectedKeywords: "k1 k2 k3", expectedCaption: "m1 m2 m3"},
		{command: "\"We\nlike\nnew\nlines\" \"yes\nwe\ndo\"", expectedError: false, expectedKeywords: "We\nlike\nnew\nlines", expectedCaption: "yes\nwe\ndo"},
		{command: "\"Unicode supporté\\? ça c'est fort\" \"héhéhé !\"", expectedError: false, expectedKeywords: "Unicode supporté\\? ça c'est fort", expectedCaption: "héhéhé !"},
	}
	for _, testCase := range testCases {
		keywords, caption, err := parseCommandLine(testCase.command, triggerGif)
		if testCase.expectedError {
			assert.NotNil(t, err, "Testing: "+testCase.command)
		} else {
			assert.Nil(t, err, "Testing: "+testCase.command)
		}
		assert.Equal(t, testCase.expectedKeywords, keywords, "Testing: "+testCase.command)
		assert.Equal(t, testCase.expectedCaption, caption, "Testing: "+testCase.command)
	}
}
