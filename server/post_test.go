package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

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
	assert.True(t, strings.Contains(err.DetailedError, errorMessage))
}
