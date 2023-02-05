package provider

import (
	"net/http"
	"testing"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultTenorResponseBody = `{
	"results": [
	  {
		"id": "4242424242",
		"title": "",
		"media_formats": {
		  "tinygifpreview": {
			"url": "https://fakeurl/tinygifpreview",
			"duration": 0,
			"preview": "",
			"dims": [
			  220,
			  220
			],
			"size": 42
		  },
		  "mp4": {
			"url": "https://fakeurl/mp4",
			"duration": 3.3,
			"preview": "",
			"dims": [
			  640,
			  640
			],
			"size": 42
		  },
		  "tinywebm": {
			"url": "https://fakeurl/tinywebm",
			"duration": 0,
			"preview": "",
			"dims": [
			  232,
			  232
			],
			"size": 42
		  },
		  "webm": {
			"url": "https://fakeurl/webm",
			"duration": 0,
			"preview": "",
			"dims": [
			  640,
			  640
			],
			"size": 42
		  },
		  "tinymp4": {
			"url": "https://fakeurl/tinymp4",
			"duration": 3.4,
			"preview": "",
			"dims": [
			  232,
			  232
			],
			"size": 42
		  },
		  "nanogif": {
			"url": "https://fakeurl/nanogif",
			"duration": 0,
			"preview": "",
			"dims": [
			  90,
			  90
			],
			"size": 42
		  },
		  "gifpreview": {
			"url": "https://fakeurl/gifpreview",
			"duration": 0,
			"preview": "",
			"dims": [
			  640,
			  640
			],
			"size": 42
		  },
		  "nanomp4": {
			"url": "https://fakeurl/nanomp4",
			"duration": 3.4,
			"preview": "",
			"dims": [
			  120,
			  120
			],
			"size": 42
		  },
		  "gif": {
			"url": "https://fakeurl/gif",
			"duration": 0,
			"preview": "",
			"dims": [
			  498,
			  498
			],
			"size": 42
		  },
		  "tinygif": {
			"url": "https://fakeurl/tinygif",
			"duration": 0,
			"preview": "",
			"dims": [
			  220,
			  220
			],
			"size": 42
		  },
		  "nanogifpreview": {
			"url": "https://fakeurl/nanogifpreview",
			"duration": 0,
			"preview": "",
			"dims": [
			  90,
			  90
			],
			"size": 42
		  },
		  "mediumgif": {
			"url": "https://fakeurl/mediumgif",
			"duration": 0,
			"preview": "",
			"dims": [
			  640,
			  640
			],
			"size": 42
		  },
		  "loopedmp4": {
			"url": "https://fakeurl/loopedmp4",
			"duration": 3.3,
			"preview": "",
			"dims": [
			  640,
			  640
			],
			"size": 42
		  },
		  "nanowebm": {
			"url": "https://fakeurl/nanowebm",
			"duration": 0,
			"preview": "",
			"dims": [
			  120,
			  120
			],
			"size": 42
		  }
		},
		"content_description": "some content description",
		"url": "https://fakeurl/fake.gif",
		"tags": [],
		"flags": [],
		"hasaudio": false
	  }
	],
	"next": "some-guid"
  }`

const (
	testTenorAPIKey    = "apikey"
	testTenorLanguage  = "fr"
	testTenorRating    = "R"
	testTenorRendition = "mediumgif"
)

func TestNewTenorProvider(t *testing.T) {
	testtHTTPClient := NewMockHTTPClient(newServerResponseOK(defaultTenorResponseBody))
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
		expectedRating      string
	}{
		{testLabel: "OK", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: testTenorRating, paramRendition: testTenorRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "KO missing rendition", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: testTenorRating, paramRendition: "", expectedError: true},
		{testLabel: "OK empty rating (legacy: empty string)", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: "", paramRendition: testTenorRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "OK empty rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: "none", paramRendition: testTenorRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "OK r rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: "r", paramRendition: testTenorRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "OK g rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: "g", paramRendition: testTenorRendition, expectedError: false, expectedRating: "high"},
		{testLabel: "OK pg rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: "pg", paramRendition: testTenorRendition, expectedError: false, expectedRating: "medium"},
		{testLabel: "OK pg-13 rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: "pg-13", paramRendition: testTenorRendition, expectedError: false, expectedRating: "low"},
		{testLabel: "OK empty language", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: "", paramRating: testTenorRating, paramRendition: testTenorRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "KO empty api key", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: "", paramLanguage: testTenorLanguage, paramRating: testTenorRating, paramRendition: testTenorRendition, expectedError: true, expectedRating: "off"},
		{testLabel: "KO nil errorGenerator", paramHTTPClient: testtHTTPClient, paramErrorGenerator: nil, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: testTenorRating, paramRendition: testTenorRendition, expectedError: true, expectedRating: "off"},
		{testLabel: "KO nil httpClient", paramHTTPClient: nil, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: testTenorRating, paramRendition: testTenorRendition, expectedError: true, expectedRating: "off"},
		{testLabel: "KO all empty", paramHTTPClient: nil, paramErrorGenerator: nil, paramAPIKey: "", paramLanguage: "", paramRating: "", paramRendition: "", expectedError: true},
	}

	for _, testCase := range testCases {
		provider, err := NewTenorProvider(testCase.paramHTTPClient, testCase.paramErrorGenerator, testCase.paramAPIKey, testCase.paramLanguage, testCase.paramRating, testCase.paramRendition)
		if testCase.expectedError {
			assert.NotNil(t, err, testCase.testLabel)
			assert.Nil(t, provider, testCase.testLabel)
		} else {
			assert.Nil(t, err, testCase.testLabel)
			assert.NotNil(t, provider, testCase.testLabel)
			assert.IsType(t, &tenor{}, provider, testCase.testLabel)
			assert.Equal(t, testCase.paramHTTPClient, provider.(*tenor).httpClient, testCase.testLabel)
			assert.Equal(t, testCase.paramErrorGenerator, provider.(*tenor).errorGenerator, testCase.testLabel)
			assert.Equal(t, testCase.paramAPIKey, provider.(*tenor).apiKey, testCase.testLabel)
			assert.Equal(t, testCase.paramLanguage, provider.(*tenor).language, testCase.testLabel)
			assert.Equal(t, testCase.expectedRating, provider.(*tenor).rating, testCase.testLabel)
			assert.Equal(t, testCase.paramRendition, provider.(*tenor).rendition, testCase.testLabel)
		}
	}
}

func generateTenorProviderForTest(mockHTTPResponse *http.Response) *tenor {
	provider, _ := NewTenorProvider(NewMockHTTPClient(mockHTTPResponse), test.MockErrorGenerator(), testTenorAPIKey, testTenorLanguage, testTenorRating, testTenorRendition)
	return provider.(*tenor)
}

func TestTenorProviderGetGifURLShouldReturnUrlWhenSearchSucceeds(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK(defaultTenorResponseBody))
	p.rendition = "tinygif"
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "https://fakeurl/tinygif")
}

func TestTenorProviderGetGifURLShouldFailIfSearchBodyIsEmpty(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK(""))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLShouldFailWhenParseError(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK("This is not a valid JSON response"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLShouldReturnEmptyUrlWhenSearchReturnNoResult(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK("{ \"weburl\": \"https://fakeurl/casdfsdfsdfsdfsdfst-gifs\", \"results\": [], \"next\": \"0\" }"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLShouldFailWhenNoURLForRendition(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK(defaultTenorResponseBody))
	p.rendition = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found for display style")
	assert.Contains(t, err.Error(), p.rendition)
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLShouldFailWhenSearchBadStatusWithoutMessage(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p := generateTenorProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLShouldFailWhenSearchBadStatusWithMessage(t *testing.T) {
	serverResponse := newServerResponseKOWithBody(429, "{ \"error\": \"Please use a registered API Key\" }")
	p := generateTenorProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Contains(t, err.Error(), "Please use a registered API Key")
	assert.Empty(t, url)
}

func generateTenorProviderForURLBuildingTests() (*tenor, *MockHTTPClient, string) {
	serverResponse := newServerResponseOK(defaultTenorResponseBody)
	client := NewMockHTTPClient(serverResponse)
	provider, _ := NewTenorProvider(client, test.MockErrorGenerator(), testTenorAPIKey, testTenorLanguage, testTenorRating, testTenorRendition)
	return provider.(*tenor), client, ""
}

func TestTenorProviderGetGifURLShouldApplyRatingFilterWhenSet(t *testing.T) {
	p, client, cursor := generateTenorProviderForURLBuildingTests()
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=off")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderGetGifURLShouldApplyLanguageFilterWhenUnset(t *testing.T) {
	p, client, cursor := generateTenorProviderForURLBuildingTests()
	p.language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "locale")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderGetGifURLShouldApplyLanguageFilterWhenSet(t *testing.T) {
	p, client, cursor := generateTenorProviderForURLBuildingTests()
	p.language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "locale="+p.language)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderGetGifURLShouldAddRandomOptionWhenRequired(t *testing.T) {
	p, client, cursor := generateTenorProviderForURLBuildingTests()
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "random=true")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, true)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
