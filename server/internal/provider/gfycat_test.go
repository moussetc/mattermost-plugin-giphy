package provider

import (
	"net/http"
	"testing"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultGfycatResponseBody = "{ \"cursor\": \"nextCursor\", \"gfycats\" : [ { \"gifUrl\": \"\", \"gif100px\": \"url0\"}, { \"gifUrl\": \"\", \"gif100px\": \"url1\"}, { \"gifUrl\": \"\", \"gif100px\": \"url2\"} ] }"

const testGfycatRendition = "gif100px"

func TestNewGfycatProvider(t *testing.T) {
	testtHTTPClient := NewMockHTTPClient(newServerResponseOK(defaultGfycatResponseBody))
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
			assert.IsType(t, &gfycat{}, provider)
			assert.Equal(t, testCase.paramHTTPClient, provider.(*gfycat).httpClient)
			assert.Equal(t, testCase.paramErrorGenerator, provider.(*gfycat).errorGenerator)
			assert.Equal(t, testCase.paramRendition, provider.(*gfycat).rendition)
		}
	}
}

func TestGfycatProviderGetGifURLShouldReturnUrlWhenSearchSucceeds(t *testing.T) {
	p, _ := NewGfycatProvider(NewMockHTTPClient(newServerResponseOK(defaultGfycatResponseBody)), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, "url0", url)
}

func TestGfycatProviderGetGifURLShouldFailIfSearchBodyIsEmpty(t *testing.T) {
	p, _ := NewGfycatProvider(NewMockHTTPClient(newServerResponseOK("")), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestGfycatProviderGetGifURLShouldFailWhenParseError(t *testing.T) {
	p, _ := NewGfycatProvider(NewMockHTTPClient(newServerResponseOK("Hello world")), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestGfycatProviderGetGifURLShouldReturnEmptyUrlWhenSearchReturnNoResult(t *testing.T) {
	p, _ := NewGfycatProvider(NewMockHTTPClient(newServerResponseOK("{\"data\": [] }")), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.Empty(t, url)
}

func TestGfycatProviderGetGifURLShouldFailWhenNoURLForRendition(t *testing.T) {
	badRendition := "NotExistingDisplayStyle"
	p, _ := NewGfycatProvider(NewMockHTTPClient(newServerResponseOK(defaultGfycatResponseBody)), test.MockErrorGenerator(), badRendition)

	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found")
	assert.Contains(t, err.Error(), badRendition)
	assert.Empty(t, url)
}

func TestGfycatProviderGetGifURLShouldFailWhenSearchBadStatus(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p, _ := NewGfycatProvider(NewMockHTTPClient(serverResponse), test.MockErrorGenerator(), testGfycatRendition)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func generateGfycatProviderForURLBuildingTests(respondeBody string) (p GifProvider, client *MockHTTPClient, cursor string) {
	serverResponse := newServerResponseOK(respondeBody)
	client = NewMockHTTPClient(serverResponse)
	p, _ = NewGfycatProvider(client, test.MockErrorGenerator(), testGfycatRendition)
	cursor = ""
	return p, client, cursor
}

func TestGfycatProviderGetGifURLWhenStartingSearch(t *testing.T) {
	p, client, cursor := generateGfycatProviderForURLBuildingTests(defaultGfycatResponseBody)

	// Check: API cursor should not be passed as initial cursor is unset
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "cursor")
		return true
	}
	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "url0", url)
	assert.Equal(t, "{\"cursorForPage\":\"\",\"positionInPage\":1}", cursor)
}

func TestGfycatProviderGetGifURLWhenCursorIsInCurrentPage(t *testing.T) {
	p, client, _ := generateGfycatProviderForURLBuildingTests(defaultGfycatResponseBody)
	cursor := "{\"cursorForPage\":\"currentCursor\",\"positionInPage\":1}"

	// Check: API cursor must be set
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "cursor=currentCursor")
		return true
	}

	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "url1", url)
	assert.Equal(t, "{\"cursorForPage\":\"currentCursor\",\"positionInPage\":2}", cursor)
}

func TestGfycatProviderGetGifURLWhenNextCursorIsDifferentPage(t *testing.T) {
	p, _, _ := generateGfycatProviderForURLBuildingTests(defaultGfycatResponseBody)
	cursor := "{\"cursorForPage\":\"currentCursor\",\"positionInPage\":2}"

	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.Equal(t, "url2", url)
	assert.Equal(t, "{\"cursorForPage\":\"nextCursor\",\"positionInPage\":0}", cursor)
}

func TestGfycatProviderGetGifURLWhenThisIsTheLastGifResult(t *testing.T) {
	p, _, cursor := generateGfycatProviderForURLBuildingTests("{ \"cursor\": \"\", \"gfycats\" : [ { \"gifUrl\": \"\", \"gif100px\": \"url0\"}] }")

	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.Equal(t, "url0", url)
	assert.Equal(t, "", cursor)
}
