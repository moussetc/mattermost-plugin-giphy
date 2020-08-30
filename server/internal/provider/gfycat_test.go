package provider

import (
	"net/http"
	"testing"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultGfycatResponseBody = "{ \"cursor\": \"mockCursor\", \"gfycats\" : [ { \"gifUrl\": \"\", \"gif100px\": \"url\"} ] }"

func generateMockConfigForGfycatProvider() pluginConf.Configuration {
	return pluginConf.Configuration{
		RenditionGfycat: "gif100px",
	}
}

func TestGfycatProviderGetGifURLOK(t *testing.T) {
	p := NewGfycatProvider(NewMockHttpClient(newServerResponseOK(defaultGfycatResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForGfycatProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "url")
}

func TestGfycatProviderGetGifURLEmptyBody(t *testing.T) {
	p := NewGfycatProvider(NewMockHttpClient(newServerResponseOK("")), test.MockErrorGenerator())
	config := generateMockConfigForGfycatProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestGfycatProviderGetGifURLParseError(t *testing.T) {
	p := NewGfycatProvider(NewMockHttpClient(newServerResponseOK("Hello world")), test.MockErrorGenerator())
	config := generateMockConfigForGfycatProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestGfycatProviderEmptyGIFList(t *testing.T) {
	p := NewGfycatProvider(NewMockHttpClient(newServerResponseOK("{\"data\": [] }")), test.MockErrorGenerator())
	config := generateMockConfigForGfycatProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No more GIF result")
	assert.Empty(t, url)
}

func TestGfycatProviderEmptyURLForRendition(t *testing.T) {
	p := NewGfycatProvider(NewMockHttpClient(newServerResponseOK(defaultGfycatResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForGfycatProvider()
	config.RenditionGfycat = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found")
	assert.Contains(t, err.Error(), config.RenditionGfycat)
	assert.Empty(t, url)
}

func TestGfycatProviderErrorStatusResponse(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p := NewGfycatProvider(NewMockHttpClient(serverResponse), test.MockErrorGenerator())
	config := generateMockConfigForGfycatProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func generateHTTPClientForGfycatParameterTest() (p GifProvider, client *MockHttpClient, config pluginConf.Configuration, cursor string) {
	serverResponse := newServerResponseOK(defaultGfycatResponseBody)
	client = NewMockHttpClient(serverResponse)
	p = NewGfycatProvider(client, test.MockErrorGenerator())
	config = generateMockConfigForGfycatProvider()
	cursor = ""
	return p, client, config, cursor
}

func TestGfycatProviderParameterCursorEmpty(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForGfycatParameterTest()

	// Cursor : optional
	// Empty initial value
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "cursor")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "mockCursor", cursor)
}

func TestGfycatProviderParameterCursorZero(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForGfycatParameterTest()

	// Initial value
	cursor = "sdfjhsdjk"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "cursor="+cursor)
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "mockCursor", cursor)
}
