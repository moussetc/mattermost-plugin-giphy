package provider

import (
	"net/http"
	"testing"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultGfycatResponseBody = "{ \"cursor\": \"mockCursor\", \"gfycats\" : [ { \"gifUrl\": \"\", \"gif100px\": \"url\"} ] }"

const testGfycatRendition = "gif100px"

func TestNewGfycatProvider(t *testing.T) {
	testtHTTPClient := NewMockHttpClient(newServerResponseOK(defaultGfycatResponseBody))
	testErrorGenerator := test.MockErrorGenerator()
	testCases := []struct {
		testLabel           string
		paramHTTPClient     HTTPClient
		paramErrorGenerator pluginError.PluginError
		paramRendition      string
		expectedError       bool
	}{
		{testLabel: "OK", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramRendition: "gif100px", expectedError: false},
		{testLabel: "KO empty rendition", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramRendition: "", expectedError: true},
		{testLabel: "KO nil errorGenerator", paramHTTPClient: testtHTTPClient, paramErrorGenerator: nil, paramRendition: "gif100px", expectedError: true},
		{testLabel: "KO nil httpClient", paramHTTPClient: nil, paramErrorGenerator: testErrorGenerator, paramRendition: "gif100px", expectedError: true},
	}

	for _, testCase := range testCases {
		provider, err := NewGfycatProvider(testCase.paramHTTPClient, testCase.paramErrorGenerator, testCase.paramRendition)
		if testCase.expectedError {
			assert.NotNil(t, err, testCase.testLabel)
			assert.Nil(t, provider, testCase.testLabel)
		} else {
			assert.Nil(t, err, testCase.testLabel)
			assert.NotNil(t, provider, testCase.testLabel)
			assert.IsType(t, &Gfycat{}, provider)
			assert.Equal(t, testCase.paramHTTPClient, provider.(*Gfycat).httpClient)
			assert.Equal(t, testCase.paramErrorGenerator, provider.(*Gfycat).errorGenerator)
			assert.Equal(t, testCase.paramRendition, provider.(*Gfycat).rendition)
		}
	}
}

func TestGfycatProviderGetGifURLOK(t *testing.T) {
	p, err := NewGfycatProvider(NewMockHttpClient(newServerResponseOK(defaultGfycatResponseBody)), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "url")
}

func TestGfycatProviderGetGifURLEmptyBody(t *testing.T) {
	p, err := NewGfycatProvider(NewMockHttpClient(newServerResponseOK("")), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestGfycatProviderGetGifURLParseError(t *testing.T) {
	p, err := NewGfycatProvider(NewMockHttpClient(newServerResponseOK("Hello world")), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestGfycatProviderEmptyGIFList(t *testing.T) {
	p, err := NewGfycatProvider(NewMockHttpClient(newServerResponseOK("{\"data\": [] }")), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No more GIF result")
	assert.Empty(t, url)
}

func TestGfycatProviderEmptyURLForRendition(t *testing.T) {
	badRendition := "NotExistingDisplayStyle"
	p, err := NewGfycatProvider(NewMockHttpClient(newServerResponseOK(defaultGfycatResponseBody)), test.MockErrorGenerator(), badRendition)

	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found")
	assert.Contains(t, err.Error(), badRendition)
	assert.Empty(t, url)
}

func TestGfycatProviderErrorStatusResponse(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p, err := NewGfycatProvider(NewMockHttpClient(serverResponse), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func generateGfycatProviderForURLBuildingTests() (p GifProvider, client *MockHttpClient, cursor string) {
	serverResponse := newServerResponseOK(defaultGfycatResponseBody)
	client = NewMockHttpClient(serverResponse)
	p, _ = NewGfycatProvider(client, test.MockErrorGenerator(), testGfycatRendition)
	cursor = ""
	return p, client, cursor
}

func TestGfycatProviderParameterCursorEmpty(t *testing.T) {
	p, client, cursor := generateGfycatProviderForURLBuildingTests()

	// Cursor : optional
	// Empty initial value
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "cursor")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "mockCursor", cursor)
}

func TestGfycatProviderParameterCursorZero(t *testing.T) {
	p, client, cursor := generateGfycatProviderForURLBuildingTests()

	// Initial value
	cursor = "sdfjhsdjk"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "cursor="+cursor)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "mockCursor", cursor)
}
