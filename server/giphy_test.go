package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const defaultGiphyResponseBody = "{\"data\" : [ { \"images\": { \"fixed_height_small\": {\"url\": \"url\"}}} ] }"

type MockHttpClient struct {
	response            *http.Response
	testRequestFunc     func(*http.Request) bool
	lastRequestPassTest bool
}

func (c *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if c.testRequestFunc != nil {
		c.lastRequestPassTest = c.testRequestFunc(req)
	}
	return c.response, nil
}

func (c *MockHttpClient) Get(s string) (*http.Response, error) {
	return c.response, nil
}

func NewMockHttpClient(res *http.Response) *MockHttpClient {
	return &MockHttpClient{
		response:        res,
		testRequestFunc: nil,
	}
}

func TestGiphyProviderGetGIFURLOK(t *testing.T) {
	p := &giphyProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(defaultGiphyResponseBody)),
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

func TestGiphyProviderMissingAPIKey(t *testing.T) {
	p := &giphyProvider{}
	config := generateMockPluginConfig()
	config.APIKey = ""
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, url)
}

func TestGiphyProviderGetGIFURLEmptyBody(t *testing.T) {
	p := &giphyProvider{}
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

func TestGiphyProviderGetGIFURLParseError(t *testing.T) {
	p := &giphyProvider{}
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

func TestGiphyProviderEmptyGIFList(t *testing.T) {
	p := &giphyProvider{}
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
	assert.Contains(t, err.Error(), "No more GIF")
	assert.Empty(t, url)
}

func TestGiphyProviderEMptyURLForRendition(t *testing.T) {
	p := &giphyProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(defaultGiphyResponseBody)),
		StatusCode: 200,
		Status:     "200 OK",
	}
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	config.Rendition = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found for display style")
	assert.Contains(t, err.Error(), config.Rendition)
	assert.Empty(t, url)
}

func TestGiphyProviderErrorStatusResponse(t *testing.T) {
	p := &giphyProvider{}
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

func TestGiphyProviderTooManyRequestStatusResponse(t *testing.T) {
	p := &giphyProvider{}
	serverResponse := &http.Response{}
	serverResponse.StatusCode = 429
	serverResponse.Status = "429 Too many requests"
	getGifProviderHttpClient = func() HttpClient { return NewMockHttpClient(serverResponse) }
	config := generateMockPluginConfig()
	cursor := ""
	url, err := p.getGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Contains(t, err.Error(), "default GIPHY API key")
	assert.Empty(t, url)
}

func generateHttpClientForParameterTest() (p *giphyProvider, client *MockHttpClient, config configuration, cursor string) {
	p = &giphyProvider{}
	serverResponse := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(defaultGiphyResponseBody)),
		StatusCode: 200,
		Status:     "200 OK",
	}
	client = NewMockHttpClient(serverResponse)
	getGifProviderHttpClient = func() HttpClient { return client }
	config = generateMockPluginConfig()
	cursor = ""
	return p, client, config, cursor
}

func TestGiphyProviderParameterAPIKey(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

	// API Key: mandatory
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "api_key")
		assert.Contains(t, req.URL.RawQuery, config.APIKey)
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterCursorEmpty(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

	// Cursor : optional
	// Empty initial value
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "offset")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterCursorZero(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

	// Initial value : 0
	cursor = "0"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "offset=0")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterCursorNotANumber(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

	// Initial value : not a number, that should be ignored
	cursor = "hahaha"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, "offset", req.URL.RawQuery)
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterRatingEmpty(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

	config.Rating = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "rating")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterRatingProvided(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

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

func TestGiphyProviderParameterLanguageEmpty(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

	config.Language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "lang")
		return true
	}
	_, err := p.getGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterLanguageProvided(t *testing.T) {
	p, client, config, cursor := generateHttpClientForParameterTest()

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
