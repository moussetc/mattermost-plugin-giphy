package provider

import (
	"net/http"
	"testing"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultGiphyResponseBodyForSearch = "{\"data\" : [ { \"images\": { \"fixed_height_small\": {\"url\": \"url\"}}} ] }"
const defaultGiphyResponseBodyForRandom = "{\"data\" : { \"images\": { \"fixed_height_small\": {\"url\": \"url\"}}} }"
const (
	testGiphyAPIKey    = "apikey"
	testGiphyLanguage  = "fr"
	testGiphyRating    = "R"
	testGiphyRendition = "fixed_height_small"
	testRootURL        = "/test"
)

func TestNewGiphyProvider(t *testing.T) {
	testtHTTPClient := NewMockHTTPClient(newServerResponseOK(defaultGiphyResponseBodyForSearch))
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
		{testLabel: "OK empty rating (legacy: empty string)", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: "", paramRendition: testGiphyRendition, expectedError: false},
		{testLabel: "OK empty rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: "none", paramRendition: testGiphyRendition, expectedError: false},
		{testLabel: "OK empty language", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: "", paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: false},
		{testLabel: "KO empty api key", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: "", paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: true},
		{testLabel: "KO nil errorGenerator", paramHTTPClient: testtHTTPClient, paramErrorGenerator: nil, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: true},
		{testLabel: "KO nil httpClient", paramHTTPClient: nil, paramErrorGenerator: testErrorGenerator, paramAPIKey: testGiphyAPIKey, paramLanguage: testGiphyLanguage, paramRating: testGiphyRating, paramRendition: testGiphyRendition, expectedError: true},
		{testLabel: "KO all empty", paramHTTPClient: nil, paramErrorGenerator: nil, paramAPIKey: "", paramLanguage: "", paramRating: "", paramRendition: "", expectedError: true},
	}

	for _, testCase := range testCases {
		provider, err := NewGiphyProvider(testCase.paramHTTPClient, testCase.paramErrorGenerator, testCase.paramAPIKey, testCase.paramLanguage, testCase.paramRating, testCase.paramRendition, testRootURL)
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
	provider, _ := NewGiphyProvider(NewMockHTTPClient(mockHTTPResponse), test.MockErrorGenerator(), testGiphyAPIKey, testGiphyLanguage, testGiphyRating, testGiphyRendition, testRootURL)
	return provider.(*giphy)
}

func TestGiphyProviderGetGifURLShouldHandleAPIErrors(t *testing.T) {
	testCases := []struct {
		testLabel     string
		httpResponse  *http.Response
		cursor        string
		expectedError string
	}{
		{testLabel: "KO empty HTTP response body", httpResponse: newServerResponseOK(""), expectedError: "empty"},
		{testLabel: "KO HTTP response JSON parse error", httpResponse: newServerResponseOK("This is not a valid JSON response"), expectedError: "parse"},
		{testLabel: "KO HTTP 400 Bad request", httpResponse: newServerResponseKO(400), expectedError: "400"},
		{testLabel: "KO HTTP 429 Too many requests", httpResponse: newServerResponseKO(429), expectedError: "default Giphy API key"},
	}

	for _, random := range [2]bool{true, false} {
		for _, testCase := range testCases {
			p := generateGiphyProviderForTest(testCase.httpResponse)
			url, err := p.GetGifURL("cat", &testCase.cursor, random)
			assert.NotNil(t, err, testCase.testLabel)
			assert.Contains(t, err.Error(), testCase.expectedError, testCase.testLabel)
			assert.Empty(t, url, testCase.testLabel)
		}
	}
}

func generateSearchAndRandomTestCases(apiResponseForSearch string, apiResponseForRandom string) (testCases []struct {
	random       bool
	httpResponse *http.Response
	label        string
}) {
	testCases = []struct {
		random       bool
		httpResponse *http.Response
		label        string
	}{
		{label: "search", random: false, httpResponse: newServerResponseOK(apiResponseForSearch)},
		{label: "random", random: true, httpResponse: newServerResponseOK(apiResponseForRandom)},
	}
	return testCases
}

func TestGiphyProviderGetGifURLShouldReturnUrlWhenSearchSucceeds(t *testing.T) {
	cursor := ""
	for _, testCase := range generateSearchAndRandomTestCases(defaultGiphyResponseBodyForSearch, defaultGiphyResponseBodyForRandom) {
		p := generateGiphyProviderForTest(testCase.httpResponse)
		url, err := p.GetGifURL("cat", &cursor, testCase.random)
		assert.Nil(t, err, testCase.label)
		assert.NotEmpty(t, url, testCase.label)
		assert.Equal(t, []string{"url"}, url, testCase.label)
	}
}

func TestGiphyProviderGetGifURLShouldReturnEmptyUrlWhenSearchReturnNoResult(t *testing.T) {
	cursor := ""

	for _, testCase := range generateSearchAndRandomTestCases("{\"data\": [] }", "{\"data\": [] }") {
		p := generateGiphyProviderForTest(testCase.httpResponse)
		url, err := p.GetGifURL("cat", &cursor, testCase.random)
		assert.Nil(t, err, testCase.label)
		assert.Empty(t, url, testCase.label)
	}
}

func TestGiphyProviderGetGifURLShouldFailWhenNoURLForRendition(t *testing.T) {
	cursor := ""
	for _, testCase := range generateSearchAndRandomTestCases(defaultGiphyResponseBodyForSearch, defaultGiphyResponseBodyForRandom) {
		p := generateGiphyProviderForTest(testCase.httpResponse)
		p.rendition = "unknown_rendition_style"
		url, err := p.GetGifURL("cat", &cursor, testCase.random)
		assert.NotNil(t, err, testCase.label)
		if testCase.random {
			assert.Contains(t, err.Error(), "No URL found for display style", testCase.label)
		} else {
			assert.Contains(t, err.Error(), "No gifs found for display style", testCase.label)
		}
		assert.Contains(t, err.Error(), p.rendition, testCase.label)
		assert.Empty(t, url, testCase.label)
	}
}

func generateGiphyProviderForURLBuildingTests(random bool) (*giphy, *MockHTTPClient, string) {
	var serverResponse *http.Response
	if random {
		serverResponse = newServerResponseOK(defaultGiphyResponseBodyForRandom)
	} else {
		serverResponse = newServerResponseOK(defaultGiphyResponseBodyForSearch)
	}
	client := NewMockHTTPClient(serverResponse)
	provider, _ := NewGiphyProvider(client, test.MockErrorGenerator(), testGiphyAPIKey, testGiphyLanguage, testGiphyRating, testGiphyRendition, testRootURL)
	return provider.(*giphy), client, ""
}

func TestGiphyProviderGetGifURLShouldUseSearchAPIWhenNotConfiguredForRandom(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests(false)

	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.Path, "/search")
		assert.Contains(t, req.URL.RawQuery, "q=cat")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderGetGifURLShouldUseRandomAPIWhenConfiguredForRandom(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests(true)

	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.Path, "/random")
		assert.Contains(t, req.URL.RawQuery, "tag=cat")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, true)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderGetGifURLUsesParameterAPIKey(t *testing.T) {
	for _, testCase := range generateSearchAndRandomTestCases(defaultGiphyResponseBodyForSearch, defaultGiphyResponseBodyForRandom) {
		p, client, cursor := generateGiphyProviderForURLBuildingTests(testCase.random)

		// API Key: mandatory
		client.testRequestFunc = func(req *http.Request) bool {
			assert.Contains(t, req.URL.RawQuery, "api_key="+testGiphyAPIKey)
			return true
		}
		_, err := p.GetGifURL("cat", &cursor, testCase.random)
		assert.Nil(t, err, testCase.label)
		assert.True(t, client.lastRequestPassTest, testCase.label)
	}
}

func TestGiphyProviderGetGifURLWhenNoRandomAndCursorIsEmpty(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests(false)

	// Cursor : optional
	// Empty initial value
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "offset")
		return true
	}

	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderGetGifURLWhenNoRandomAndCursorIsZero(t *testing.T) {
	p, client, _ := generateGiphyProviderForURLBuildingTests(false)

	// Initial value : 0
	cursor := "0"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "offset=0")
		return true
	}

	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderGetGifURLWhenNoRandomAndCursorIsNotANumber(t *testing.T) {
	p, client, _ := generateGiphyProviderForURLBuildingTests(false)

	// Initial value : not a number, that should be ignored
	cursor := "hahaha"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, "offset", req.URL.RawQuery)
		return true
	}

	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
	assert.Equal(t, "1", cursor)
}

func TestGiphyProviderGetGifURLShouldApplyRatingFilterWhenUnset(t *testing.T) {
	for _, testCase := range generateSearchAndRandomTestCases(defaultGiphyResponseBodyForSearch, defaultGiphyResponseBodyForRandom) {
		p, client, cursor := generateGiphyProviderForURLBuildingTests(testCase.random)
		p.rating = ""
		client.testRequestFunc = func(req *http.Request) bool {
			assert.NotContains(t, req.URL.RawQuery, "rating")
			return true
		}

		_, err := p.GetGifURL("cat", &cursor, testCase.random)
		assert.Nil(t, err, testCase.label)
		assert.True(t, client.lastRequestPassTest, testCase.label)
	}
}

func TestGiphyProviderGetGifURLShouldApplyRatingFilterWhenNoRandomAndRatingFilter(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests(false)
	p.rating = "RATING"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "rating="+p.rating)
		return true
	}

	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderGetGifURLShouldNotApplyLanguageFilterWhenNoRandomAndNoLanguageSet(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests(false)
	p.language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "lang")
		return true
	}

	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestGiphyProviderGetGifURLShouldApplyLanguageFilterWhenNoRandomAndLanguageSet(t *testing.T) {
	p, client, cursor := generateGiphyProviderForURLBuildingTests(false)
	p.language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "lang="+p.language)
		return true
	}

	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
