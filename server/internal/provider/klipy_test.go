package provider

import (
	"net/http"
	"testing"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultKlipyResponseBody = `{
  "results": [
    {
      "id": "4242424242",
      "title": "",
      "media_formats": {
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
        "webp": {
          "url": "https://fakeurl/webp",
          "duration": 0,
          "preview": "",
          "dims": [
            640,
            640
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
        "tinywebm": {
          "url": "https://fakeurl/tinywebm",
          "duration": 0,
          "preview": "",
          "dims": [
            320,
            320
          ],
          "size": 42
        },
        "nanowebm": {
          "url": "https://fakeurl/nanowebm",
          "duration": 0,
          "preview": "",
          "dims": [
            150,
            150
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
        "tinymp4": {
          "url": "https://fakeurl/tinymp4",
          "duration": 3.4,
          "preview": "",
          "dims": [
            320,
            320
          ],
          "size": 42
        },
        "nanomp4": {
          "url": "https://fakeurl/nanomp4",
          "duration": 3.4,
          "preview": "",
          "dims": [
            150,
            150
          ],
          "size": 42
        }
      },
      "created": 1765483200,
      "content_description": "some content description",
      "itemurl": "https://fakeurl/gifs/greetings",
      "url": "https://fakeurl/fake.gif",
      "tags": [],
      "flags": [],
      "hasaudio": false,
      "content_description_source": ""
    }
  ],
  "next": "some-guid"
}`

const (
	testKlipyAPIKey    = "apikey"
	testKlipyLanguage  = "fr"
	testKlipyRating    = "R"
	testKlipyRendition = "mediumgif"
	testKlipyRootURL   = "/test"
)

func TestNewKlipyProvider(t *testing.T) {
	testtHTTPClient := NewMockHTTPClient(newServerResponseOK(defaultKlipyResponseBody))
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
		{testLabel: "OK", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: testKlipyRating, paramRendition: testKlipyRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "KO missing rendition", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: testKlipyRating, paramRendition: "", expectedError: true},
		{testLabel: "OK empty rating (legacy: empty string)", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: "", paramRendition: testKlipyRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "OK empty rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: "none", paramRendition: testKlipyRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "OK r rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: "r", paramRendition: testKlipyRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "OK g rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: "g", paramRendition: testKlipyRendition, expectedError: false, expectedRating: "high"},
		{testLabel: "OK pg rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: "pg", paramRendition: testKlipyRendition, expectedError: false, expectedRating: "medium"},
		{testLabel: "OK pg-13 rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: "pg-13", paramRendition: testKlipyRendition, expectedError: false, expectedRating: "low"},
		{testLabel: "OK empty language", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: "", paramRating: testKlipyRating, paramRendition: testKlipyRendition, expectedError: false, expectedRating: "off"},
		{testLabel: "KO empty api key", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: "", paramLanguage: testKlipyLanguage, paramRating: testKlipyRating, paramRendition: testKlipyRendition, expectedError: true, expectedRating: "off"},
		{testLabel: "KO nil errorGenerator", paramHTTPClient: testtHTTPClient, paramErrorGenerator: nil, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: testKlipyRating, paramRendition: testKlipyRendition, expectedError: true, expectedRating: "off"},
		{testLabel: "KO nil httpClient", paramHTTPClient: nil, paramErrorGenerator: testErrorGenerator, paramAPIKey: testKlipyAPIKey, paramLanguage: testKlipyLanguage, paramRating: testKlipyRating, paramRendition: testKlipyRendition, expectedError: true, expectedRating: "off"},
		{testLabel: "KO all empty", paramHTTPClient: nil, paramErrorGenerator: nil, paramAPIKey: "", paramLanguage: "", paramRating: "", paramRendition: "", expectedError: true},
	}

	for _, testCase := range testCases {
		provider, err := NewKlipyProvider(testCase.paramHTTPClient, testCase.paramErrorGenerator, testCase.paramAPIKey, testCase.paramLanguage, testCase.paramRating, testCase.paramRendition, testKlipyRootURL)
		if testCase.expectedError {
			assert.NotNil(t, err, testCase.testLabel)
			assert.Nil(t, provider, testCase.testLabel)
		} else {
			assert.Nil(t, err, testCase.testLabel)
			assert.NotNil(t, provider, testCase.testLabel)
			assert.IsType(t, &klipy{}, provider, testCase.testLabel)
			assert.Equal(t, testCase.paramHTTPClient, provider.(*klipy).httpClient, testCase.testLabel)
			assert.Equal(t, testCase.paramErrorGenerator, provider.(*klipy).errorGenerator, testCase.testLabel)
			assert.Equal(t, testCase.paramAPIKey, provider.(*klipy).apiKey, testCase.testLabel)
			assert.Equal(t, testCase.paramLanguage, provider.(*klipy).language, testCase.testLabel)
			assert.Equal(t, testCase.expectedRating, provider.(*klipy).rating, testCase.testLabel)
			assert.Equal(t, testCase.paramRendition, provider.(*klipy).rendition, testCase.testLabel)
		}
	}
}

func generateKlipyProviderForTest(mockHTTPResponse *http.Response) *klipy {
	provider, _ := NewKlipyProvider(NewMockHTTPClient(mockHTTPResponse), test.MockErrorGenerator(), testKlipyAPIKey, testKlipyLanguage, testKlipyRating, testKlipyRendition, testKlipyRootURL)
	return provider.(*klipy)
}

func TestKlipyProviderGetGifURLShouldReturnUrlWhenSearchSucceeds(t *testing.T) {
	p := generateKlipyProviderForTest(newServerResponseOK(defaultKlipyResponseBody))
	p.rendition = "tinygif"
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, []string{"https://fakeurl/tinygif"})
}

func TestKlipyProviderGetGifURLShouldFailIfSearchBodyIsEmpty(t *testing.T) {
	p := generateKlipyProviderForTest(newServerResponseOK(""))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestKlipyProviderGetGifURLShouldFailWhenParseError(t *testing.T) {
	p := generateKlipyProviderForTest(newServerResponseOK("This is not a valid JSON response"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestKlipyProviderGetGifURLShouldReturnEmptyUrlWhenSearchReturnNoResult(t *testing.T) {
	p := generateKlipyProviderForTest(newServerResponseOK("{ \"weburl\": \"https://fakeurl/casdfsdfsdfsdfsdfst-gifs\", \"results\": [], \"next\": \"0\" }"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.Empty(t, url)
}

func TestKlipyProviderGetGifURLShouldFailWhenNoURLForRendition(t *testing.T) {
	p := generateKlipyProviderForTest(newServerResponseOK(defaultKlipyResponseBody))
	p.rendition = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No gifs found for display style")
	assert.Contains(t, err.Error(), p.rendition)
	assert.Empty(t, url)
}

func TestKlipyProviderGetGifURLShouldFailWhenSearchBadStatusWithoutMessage(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p := generateKlipyProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func TestKlipyProviderGetGifURLShouldFailWhenSearchBadStatusWithMessage(t *testing.T) {
	serverResponse := newServerResponseKOWithBody(429, "{ \"error\": \"Please use a registered API Key\" }")
	p := generateKlipyProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor, false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Contains(t, err.Error(), "Please use a registered API Key")
	assert.Empty(t, url)
}

func generateKlipyProviderForURLBuildingTests() (*klipy, *MockHTTPClient, string) {
	serverResponse := newServerResponseOK(defaultKlipyResponseBody)
	client := NewMockHTTPClient(serverResponse)
	provider, _ := NewKlipyProvider(client, test.MockErrorGenerator(), testKlipyAPIKey, testKlipyLanguage, testKlipyRating, testKlipyRendition, testKlipyRootURL)
	return provider.(*klipy), client, ""
}

func TestKlipyProviderGetGifURLShouldApplyRatingFilterWhenSet(t *testing.T) {
	p, client, cursor := generateKlipyProviderForURLBuildingTests()
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=off")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestKlipyProviderGetGifURLShouldApplyLanguageFilterWhenUnset(t *testing.T) {
	p, client, cursor := generateKlipyProviderForURLBuildingTests()
	p.language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "locale")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestKlipyProviderGetGifURLShouldApplyLanguageFilterWhenSet(t *testing.T) {
	p, client, cursor := generateKlipyProviderForURLBuildingTests()
	p.language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "locale="+p.language)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, false)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestKlipyProviderGetGifURLShouldAddRandomOptionWhenRequired(t *testing.T) {
	p, client, cursor := generateKlipyProviderForURLBuildingTests()
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "random=true")
		assert.NotContains(t, req.URL.RawQuery, "limit=1")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor, true)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
