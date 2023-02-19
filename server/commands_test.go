package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest/mock"
)

var testArgs = &model.CommandArgs{
	RootId:    "42",
	ChannelId: "43",
	UserId:    testUserID,
}

func TestRegisterCommandsShouldFailWhenRegisterGifCommandFails(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(config))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGif })).Return(errors.New("fail mock register command"))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGifWithPreview })).Return(nil)
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	p := Plugin{}
	p.configuration = &config
	p.SetAPI(api)

	assert.NotNil(t, p.RegisterCommands())
}

func TestRegisterCommandsShouldFailWhenRegisterGifPreviewCommandFails(t *testing.T) {
	api := &plugintest.API{}
	config := generateMockPluginConfig()
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*configuration.Configuration")).Return(mockLoadConfig(config))
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGif })).Return(nil)
	api.On("RegisterCommand", mock.MatchedBy(func(command *model.Command) bool { return command.Trigger == config.CommandTriggerGifWithPreview })).Return(errors.New("fail mock register command"))
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)
	p := Plugin{}
	p.configuration = &config
	p.SetAPI(api)

	assert.NotNil(t, p.RegisterCommands())
}

func TestExecuteCommandGifShouldReturnInChannelResponseWhenSearchSucceeds(t *testing.T) {
	_, p := initMockAPI()
	p.gifProvider = newMockGifProvider()

	response, err := p.executeCommandGif(testKeywords, testCaption, testArgs)

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, model.CommandResponseTypeInChannel, response.ResponseType)
	assert.True(t, strings.Contains(response.Text, testKeywords))
	assert.True(t, strings.Contains(response.Text, testCaption))
}

func TestExecuteCommandGifShouldSendEphemeralPostWhenSearchReturnsNoResult(t *testing.T) {
	api, p := initMockAPI()
	p.gifProvider = &emptyGifProvider{}
	api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil)

	response, err := p.executeCommandGif(testKeywords, testCaption, testArgs)

	assert.Nil(t, err)
	assert.NotNil(t, response)
	api.AssertNumberOfCalls(t, "SendEphemeralPost", 1)
	api.AssertCalled(t, "SendEphemeralPost",
		mock.MatchedBy(func(uID string) bool {
			return uID == testArgs.UserId
		}),
		mock.MatchedBy(func(post *model.Post) bool {
			return post != nil &&
				post.ChannelId == testArgs.ChannelId &&
				post.UserId == p.botID &&
				strings.Contains(post.Message, "found") &&
				strings.Contains(post.Message, testKeywords)
		}))
}

func TestExecuteCommandGifShouldLogAndFailWhenSearchFails(t *testing.T) {
	api, p := initMockAPI()
	errorMessage := "ARGHHHH"
	p.gifProvider = &mockGifProviderFail{errorMessage}
	api.On("LogWarn", mock.AnythingOfType("string")).Return(nil)

	response, err := p.executeCommandGif("mayhem", "guy", testArgs)
	assert.NotNil(t, err)
	assert.Empty(t, response)
	assert.Contains(t, err.DetailedError, errorMessage)
	api.AssertNumberOfCalls(t, "LogWarn", 1)
}

func TestExecuteCommandGifWithPreviewShouldPostAnEphemeralGifPostWhenSearchSucceeds(t *testing.T) {
	api, p := initMockAPI()
	p.gifProvider = newMockGifProvider()

	var recordCreationPost *model.Post
	api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(nil, nil).Run(func(args mock.Arguments) {
		recordCreationPost = args.Get(1).(*model.Post)
	})

	response, err := p.executeCommandGifWithPreview(testKeywords, testCaption, testArgs)

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", response.ResponseType)
	assert.NotNil(t, recordCreationPost)
	assert.True(t, strings.Contains(recordCreationPost.Message, testKeywords))
	assert.True(t, strings.Contains(recordCreationPost.Message, testCaption))
	assert.Equal(t, recordCreationPost.RootId, testArgs.RootId)
	assert.Equal(t, recordCreationPost.ChannelId, testArgs.ChannelId)
}

func TestExecuteCommandGifWithPreviewShouldReturnEphemeralResponseWhenSearchReturnsNoResult(t *testing.T) {
	api, p := initMockAPI()
	p.gifProvider = &emptyGifProvider{}
	api.On("SendEphemeralPost", mock.Anything, mock.Anything).Return(nil)

	response, err := p.executeCommandGifWithPreview(testKeywords, testCaption, testArgs)

	assert.Nil(t, err)
	assert.NotNil(t, response)
	api.AssertNumberOfCalls(t, "SendEphemeralPost", 1)
	api.AssertCalled(t, "SendEphemeralPost",
		mock.MatchedBy(func(uID string) bool { return uID == testArgs.UserId }),
		mock.MatchedBy(func(post *model.Post) bool {
			return post != nil &&
				post.ChannelId == testArgs.ChannelId &&
				post.UserId == p.botID &&
				strings.Contains(post.Message, "found") &&
				strings.Contains(post.Message, testKeywords)
		}))
}

func TestExecuteCommandGifWithPreviewShouldLogAndFailWhenSearchFails(t *testing.T) {
	api, p := initMockAPI()
	p.gifProvider = &mockGifProviderFail{"mockError"}
	api.On("LogWarn", mock.AnythingOfType("string")).Return(nil)

	response, err := p.executeCommandGifWithPreview("hello", "", nil)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "mockError")
	assert.Nil(t, response)
	api.AssertNumberOfCalls(t, "LogWarn", 1)
}

func TestGeneratePreviewPostAttachments(t *testing.T) {
	gifURLs := []string{testGifURLPrevious, testGifURL}
	attachments := generatePreviewPostAttachments(testKeywords, testCaption, testCursor, testRootID, gifURLs, 0)

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
		assert.Equal(t, testKeywords, context[contextKeywords])
		assert.Equal(t, gifURLs, context[contextGifURLs])
		assert.Equal(t, testCursor, context[contextAPICursor])
		assert.Equal(t, testRootID, context[contextRootID])
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
