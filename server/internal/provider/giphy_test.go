package provider

import (
	"net/http"
	"testing"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultGiphyResponseBody = "{\"data\" : [ { \"images\": { \"fixed_height_small\": {\"url\": \"url\"}}} ] }"

func generateMockConfigForGiphyProvider() pluginConf.Configuration {
	return pluginConf.Configuration{
		APIKey:    "defaultAPIKey",
		Rating:    "",
		Language:  "fr",
		Rendition: "fixed_height_small",
	}
}

func TestGiphyProviderGetGifURLOK(t *testing.T) {
	p := NewGiphyProvider(NewMockHttpClient(newServerResponseOK(defaultGiphyResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "url")
}

func TestGiphyProviderMissingAPIKey(t *testing.T) {
	p := NewGiphyProvider(NewMockHttpClient(newServerResponseOK(defaultGiphyResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	config.APIKey = ""
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, url)
}

func TestGiphyProviderGetGifURLEmptyBody(t *testing.T) {
	p := NewGiphyProvider(NewMockHttpClient(newServerResponseOK("")), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestGiphyProviderGetGifURLParseError(t *testing.T) {
	p := NewGiphyProvider(NewMockHttpClient(newServerResponseOK("Hello World")), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestGiphyProviderEmptyGIFList(t *testing.T) {
	p := NewGiphyProvider(NewMockHttpClient(newServerResponseOK("{\"data\": [] }")), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No more GIF")
	assert.Empty(t, url)
}

func TestGiphyProviderEmptyURLForRendition(t *testing.T) {
	p := NewGiphyProvider(NewMockHttpClient(newServerResponseOK(defaultGiphyResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	config.Rendition = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found for display style")
	assert.Contains(t, err.Error(), config.Rendition)
	assert.Empty(t, url)
}

func TestGiphyProviderErrorStatusResponse(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p := NewGiphyProvider(NewMockHttpClient(serverResponse), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func TestGiphyProviderTooManyRequestStatusResponse(t *testing.T) {
	serverResponse := newServerResponseKO(429)
	p := NewGiphyProvider(NewMockHttpClient(serverResponse), test.MockErrorGenerator())
	config := generateMockConfigForGiphyProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Contains(t, err.Error(), "default GIPHY API key")
	assert.Empty(t, url)
}

func generateHTTPClientForParameterTest() (p GifProvider, client *MockHttpClient, config pluginConf.Configuration, cursor string) {
	serverResponse := newServerResponseOK(defaultGiphyResponseBody)
	client = NewMockHttpClient(serverResponse)
	p = NewGiphyProvider(client, test.MockErrorGenerator())
	config = generateMockConfigForGiphyProvider()
	cursor = ""
	return p, client, config, cursor
}

func TestGiphyProviderParameterAPIKey(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	// API Key: mandatory
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "api_key")
		assert.Contains(t, req.URL.RawQuery, config.APIKey)
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterCursorEmpty(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	// Cursor : optional
	// Empty initial value
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "offset")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterCursorZero(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	// Initial value : 0
	cursor = "0"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "offset=0")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterCursorNotANumber(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	// Initial value : not a number, that should be ignored
	cursor = "hahaha"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, "offset", req.URL.RawQuery)
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterRatingEmpty(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	config.Rating = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "rating")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterRatingProvided(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	// Initial value : 0
	config.Rating = "RATING"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "rating=RATING")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterLanguageEmpty(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	config.Language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "lang")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterLanguageProvided(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForParameterTest()

	// Initial value : 0
	config.Language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "lang=Moldovalaque")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
