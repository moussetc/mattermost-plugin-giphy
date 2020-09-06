package provider

import (
	"net/http"
	"testing"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultGiphyResponseBody = "{\"data\" : [ { \"images\": { \"fixed_height_small\": {\"url\": \"url\"}}} ] }"
const (
	testGiphyAPIKey    = "apikey"
	testGiphyLanguage  = "fr"
	testGiphyRating    = "R"
	testGiphyRendition = "fixed_height_small"
)

func TestNewGiphyProvider(t *testing.T) {
	testtHTTPClient := NewMockHttpClient(newServerResponseOK(defaultGiphyResponseBody))
	testErrorGenerator := test.MockErrorGenerator()
	testCases := []struct {
		testLabel           string
		paramHTTPClient     HTTPClient
		paramErrorGenerator pluginError.PluginError
		paramAPIKey         string
		paramRating         string
		paramLanguage       string
		paramRendition      string
		expectedError       bool
	}{
		{testLabel: "OK", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: false},
		{testLabel: "KO missing rendition", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: "", expectedError: true},
		{testLabel: "OK empty rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: "", paramRendition: testGiphyRendition, expectedError: false},
		{testLabel: "OK empty language", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: "", paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: false},
		{testLabel: "KO empty api key", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: "", paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: true},
		{testLabel: "KO nil errorGenerator", paramHTTPClient: testtHTTPClient, paramErrorGenerator: nil, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: true},
		{testLabel: "KO nil httpClient", paramHTTPClient: nil, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: true},
		{testLabel: "KO all empty", paramHTTPClient: nil, paramErrorGenerator: nil, paramAPIKey: "", paramLanguage: "", paramRating: "", paramRendition: "", expectedError: true},
	}

	for _, testCase := range testCases {
		provider, err := NewGiphyProvider(testCase.paramHTTPClient, testCase.paramErrorGenerator, testCase.paramAPIKey, testCase.paramLanguage, testCase.paramRating, testCase.paramRendition)
		if testCase.expectedError {
			assert.NotNil(t, err, testCase.testLabel)
			assert.Nil(t, provider, testCase.testLabel)
		} else {
			assert.Nil(t, err, testCase.testLabel)
			assert.NotNil(t, provider, testCase.testLabel)
			assert.IsType(t, &giphy{}, provider, testCase.testLabel)
			assert.Equal(t, testCase.paramHTTPClient, provider.(*giphy).httpClient, testCase.testLabel)
			assert.Equal(t, testCase.paramErrorGenerator, provider.(*giphy).errorGenerator, testCase.testLabel)
			assert.Equal(t, testCase.paramAPIKey, provider.(*giphy).apiKey, testCase.testLabel)
			assert.Equal(t, testCase.paramLanguage, provider.(*giphy).language, testCase.testLabel)
			assert.Equal(t, testCase.paramRating, provider.(*giphy).rating, testCase.testLabel)
			assert.Equal(t, testCase.paramRendition, provider.(*giphy).rendition, testCase.testLabel)
		}
	}
}

func generateGiphyProviderForTest(mockHTTPResponse *http.Response) *giphy {
	provider, _ := NewGiphyProvider(NewMockHttpClient(mockHTTPResponse), test.MockErrorGenerator(), testGiphyAPIKey, testGiphyLanguage, testGiphyRating, testGiphyRendition)
	return provider.(*giphy)
}

func TestGiphyProviderGetGifURLOK(t *testing.T) {
	p := generateGiphyProviderForTest(newServerResponseOK(defaultGiphyResponseBody))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "url")
}

func TestGiphyProviderGetGifURLEmptyBody(t *testing.T) {
	p := generateGiphyProviderForTest(newServerResponseOK(""))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestGiphyProviderGetGifURLParseError(t *testing.T) {
	p := generateGiphyProviderForTest(newServerResponseOK("This is not a valid JSON response"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestGiphyProviderEmptyGIFList(t *testing.T) {
	p := generateGiphyProviderForTest(newServerResponseOK("{\"data\": [] }"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No more GIF")
	assert.Empty(t, url)
}

func TestGiphyProviderEmptyURLForRendition(t *testing.T) {
	p := generateGiphyProviderForTest(newServerResponseOK(defaultGiphyResponseBody))
	p.rendition = "unknown_rendition_style"
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found for display style")
	assert.Contains(t, err.Error(), p.rendition)
	assert.Empty(t, url)
}

func TestGiphyProviderErrorStatusResponse(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p := generateGiphyProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func TestGiphyProviderTooManyRequestStatusResponse(t *testing.T) {
	serverResponse := newServerResponseKO(429)
	p := generateGiphyProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Contains(t, err.Error(), "default Giphy API key")
	assert.Empty(t, url)
}

func generateGiphyProviderForURLBuildingTests() (*giphy, *MockHttpClient, string) {
	serverResponse := newServerResponseOK(defaultGiphyResponseBody)
	client := NewMockHttpClient(serverResponse)
	provider, _ := NewGiphyProvider(client, test.MockErrorGenerator(), testGiphyAPIKey, testGiphyLanguage, testGiphyRating, testGiphyRendition)
	return provider.(*giphy), client, ""
}

func TestGiphyProviderParameterAPIKey(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()

	// API Key: mandatory
	client.testRequestFunc = func(req *http.Request) bool {
		//	assert.Contains(t, req.URL.RawQuery, "api_key")
		//assert.Contains(t, req.URL.RawQuery, testAPIKey)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterCursorEmpty(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()

	// Cursor : optional
	// Empty initial value
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "offset")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterCursorZero(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()

	// Initial value : 0
	cursor = "0"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "offset=0")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterCursorNotANumber(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()

	// Initial value : not a number, that should be ignored
	cursor = "hahaha"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, "offset", req.URL.RawQuery)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderParameterRatingEmpty(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()
	p.rating = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "rating")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterRatingProvided(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()
	p.rating = "RATING"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "rating="+p.rating)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterLanguageEmpty(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()
	p.language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "lang")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderParameterLanguageProvided(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests()
	p.language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "lang="+p.language)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
