package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const defaultGfycatResponseBody = "{ \"cursor\": \"mockCursor\", \"gfycats\" : [ { \"gifUrl\": \"\", \"gif100Px\": \"url\"} ] }"

func TestGfycatProviderGetGIFURLOK(t *testing.T) {
	p := &gfyCatProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(defaultGfycatResponseBody)),
		StatusCode: 200,
		Status:     "200 OK",
	}
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "url")
}

/* func TestGfycatProviderMissingAPIKey(t *testing.T) {
	p := &gfyCatProvider{}
	config := generateMockPluginConfig()
	config.APIKey = ""
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, url)
}*/

func TestGfycatProviderGetGIFURLEmptyBody(t *testing.T) {
	p := &gfyCatProvider{}
	serverResponse := &http.Response{}
	serverResponse.StatusCode = 200
	serverResponse.Status = "200 OK"
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestGfycatProviderGetGIFURLParseError(t *testing.T) {
	p := &gfyCatProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString("Hello World")),
		StatusCode: 200,
		Status:     "200 OK",
	}
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestGfycatProviderEmptyGIFList(t *testing.T) {
	p := &gfyCatProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString("{\"data\": [] }")),
		StatusCode: 200,
		Status:     "200 OK",
	}
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No more GIF result")
	assert.Empty(t, url)
}

func TestGfycatProviderEmptyURLForRendition(t *testing.T) {
	p := &gfyCatProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(defaultGfycatResponseBody)),
		StatusCode: 200,
		Status:     "200 OK",
	}
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	config.RenditionGfycat = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found")
	assert.Contains(t, err.Error(), config.RenditionGfycat)
	assert.Empty(t, url)
}

func TestGfycatProviderErrorStatusResponse(t *testing.T) {
	p := &gfyCatProvider{}
	serverResponse := &http.Response{}
	serverResponse.StatusCode = 400
	serverResponse.Status = "400 Bad Request"
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func generateHttpClientForGfycatParameterTest() (p *gfyCatProvider, client *MockHttpClient, config configuration, cursor string) {
	p = &gfyCatProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(defaultGfycatResponseBody)),
		StatusCode: 200,
		Status:     "200 OK",
	}
	client = NewMockHttpClient(serverResponse)
	getGifProviderHttpClient = func() HttpClient { return client }
	config = generateMockPluginConfig()
	cursor = ""
	return p, client, config, cursor
}

/*func TestGfycatProviderParameterAPIKey(t *testing.T) {
	p, client, config, cursor := generateHttpClientForGfycatParameterTest()

	// API Key: mandatory
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "api_key")
		assert.Contains(t, req.URL.RawQuery, config.APIKey)
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}*/

func TestGfycatProviderParameterCursorEmpty(t *testing.T) {
	p, client, config, cursor := generateHttpClientForGfycatParameterTest()

	// Cursor : optional
	// Empty initial value
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "cursor")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "mockCursor", cursor)
}

func TestGfycatProviderParameterCursorZero(t *testing.T) {
	p, client, config, cursor := generateHttpClientForGfycatParameterTest()

	// Initial value
	cursor = "sdfjhsdjk"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "cursor="+cursor)
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "mockCursor", cursor)
}

/* func TestGfycatProviderParameterRatingEmpty(t *testing.T) {
	p, client, config, cursor := generateHttpClientForGfycatParameterTest()

	config.Rating = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "rating")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGfycatProviderParameterRatingProvided(t *testing.T) {
	p, client, config, cursor := generateHttpClientForGfycatParameterTest()

	// Initial value : 0
	config.Rating = "RATING"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "rating=RATING")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGfycatProviderParameterLanguageEmpty(t *testing.T) {
	p, client, config, cursor := generateHttpClientForGfycatParameterTest()

	config.Language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "lang")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGfycatProviderParameterLanguageProvided(t *testing.T) {
	p, client, config, cursor := generateHttpClientForGfycatParameterTest()

	// Initial value : 0
	config.Language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "lang=Moldovalaque")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
*/
