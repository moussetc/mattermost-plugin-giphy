package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGiphyProviderGetGIFURL(t *testing.T) {
	p := &giphyProvider{}
	config := &GiphyPluginConfiguration{Language: "fr", Rating: "", Rendition: "fixed_height_small"}
	url, err := p.getGifURL(config, "cat")
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
}
