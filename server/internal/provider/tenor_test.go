package provider

import (
	"net/http"
	"testing"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultTenorResponseBody = "{  \"weburl\": \"https://fakeurl/search/stuff\",   \"results\": [   {    \"tags\": [],     \"url\": \"https://fakeurl/fake.gif\",     \"media\": [     {      \"nanomp4\": {       \"url\": \"https://fakeurl/nanomp4\",        \"dims\": [        150,         112       ],        \"duration\": 0.5,        \"preview\": \"https://fakeurl/preview.png\",        \"size\": 7343      },       \"nanowebm\": {       \"url\": \"https://fakeurl/nanowebm\",        \"dims\": [        150,         112       ],        \"preview\": \"https://fakeurl/nanowebm/preview.png\",        \"size\": 9550      },       \"tinygif\": {       \"url\": \"https://fakeurl/tinygif.gif\",        \"dims\": [        220,         164       ],        \"preview\": \"https://fakeurl/tinigif/preview.gif\",        \"size\": 22519      },       \"tinymp4\": {       \"url\": \"https://fakeurl/mp4\",        \"dims\": [        320,         238       ],        \"duration\": 0.5,        \"preview\": \"https://fakeurl/tinymp4/preview.png\",        \"size\": 17732      },       \"tinywebm\": {       \"url\": \"https://fakeurl/tinywebm\",        \"dims\": [        320,         238       ],        \"preview\": \"https://fakeurl/tinywebm/preview.png\",        \"size\": 12311      },       \"webm\": {       \"url\": \"https://fakeurl/webm\",        \"dims\": [        444,         332       ],        \"preview\": \"https://fakeurl/webm/preview.png\",        \"size\": 14924      },       \"gif\": {       \"url\": \"https://fakeurl/gif.gif\",        \"dims\": [        444,         332       ],        \"preview\": \"https://fakeurl/gif/preview.png\",        \"size\": 465547      },       \"mp4\": {       \"url\": \"https://fakeurl/mp4\",        \"dims\": [        444,         332       ],        \"duration\": 0.5,        \"preview\": \"https://fakeurl/mp4/preview.png\",        \"size\": 36818      },       \"loopedmp4\": {       \"url\": \"https://fakeurl/loopedmp4\",        \"dims\": [        444,         332       ],        \"duration\": 1.5,        \"preview\": \"https://fakeurl/loopedmp4/preview.png\",        \"size\": 108909      },       \"mediumgif\": {       \"url\": \"https://fakeurl/mediumgif.gif\",        \"dims\": [        444,         332       ],        \"preview\": \"https://fakeurl/imediumgif/preview.gif\",        \"size\": 127524      },       \"nanogif\": {       \"url\": \"https://fakeurl/nanogif.gif\",        \"dims\": [        120,         90       ],        \"preview\": \"https://fakeurl/nanogif/preview.gif\",        \"size\": 9104      }     }    ],     \"created\": 1476975012.524378,     \"shares\": 1,     \"itemurl\": \"https://fakeurl/view/fakeurl\",     \"composite\": null,     \"hasaudio\": false,     \"title\": \"\",     \"id\": \"6198981\"   }  ],   \"next\": \"1\" }"
const (
	testTenorAPIKey    = "apikey"
	testTenorLanguage  = "fr"
	testTenorRating    = "R"
	testTenorRendition = "mediumgif"
)

func TestNewTenorProvider(t *testing.T) {
	testtHTTPClient := NewMockHttpClient(newServerResponseOK(defaultTenorResponseBody))
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
		{testLabel: "OK empty rating", paramHTTPClient: testtHTTPClient, paramErrorGenerator: testErrorGenerator, paramAPIKey: testTenorAPIKey, paramLanguage: testTenorLanguage, paramRating: "", paramRendition: testTenorRendition, expectedError: false, expectedRating: "off"},
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
	provider, _ := NewTenorProvider(NewMockHttpClient(mockHTTPResponse), test.MockErrorGenerator(), testTenorAPIKey, testTenorLanguage, testTenorRating, testTenorRendition)
	return provider.(*tenor)
}

func generateMockConfigForTenorProvider() pluginConf.Configuration {
	return pluginConf.Configuration{
		APIKey:         "defaultAPIKey",
		Rating:         "",
		Language:       "fr",
		RenditionTenor: "mediumgif",
	}
}

func TestTenorProviderGetGifURLOK(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK(defaultTenorResponseBody))
	p.rendition = "tinygif"
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "https://fakeurl/tinygif.gif")
}

func TestTenorProviderGetGifURLEmptyBody(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK(""))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLParseError(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK("This is not a valid JSON response"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestTenorProviderEmptyGIFList(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK("{ \"weburl\": \"https://fakeurl/casdfsdfsdfsdfsdfst-gifs\", \"results\": [], \"next\": \"0\" }"))
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No more GIF result")
	assert.Empty(t, url)
}

func TestTenorProviderEmptyURLForRendition(t *testing.T) {
	p := generateTenorProviderForTest(newServerResponseOK(defaultTenorResponseBody))
	p.rendition = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found for display style")
	assert.Contains(t, err.Error(), p.rendition)
	assert.Empty(t, url)
}

func TestTenorProviderErrorStatusResponseWithoutErrorMessage(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p := generateTenorProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func TestTenorProviderErrorStatusResponseWithErrorMessage(t *testing.T) {
	serverResponse := newServerResponseKOWithBody(429, "{ \"error\": \"Please use a registered API Key\" }")
	p := generateTenorProviderForTest(serverResponse)
	cursor := ""
	url, err := p.GetGifURL("cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Contains(t, err.Error(), "Please use a registered API Key")
	assert.Empty(t, url)
}

func generatTenorProviderForURLBuildingTests() (*tenor, *MockHttpClient, string) {
	serverResponse := newServerResponseOK(defaultTenorResponseBody)
	client := NewMockHttpClient(serverResponse)
	provider, _ := NewTenorProvider(client, test.MockErrorGenerator(), testTenorAPIKey, testTenorLanguage, testTenorRating, testTenorRendition)
	return provider.(*tenor), client, ""
}

func TestTenorProviderParameterRating(t *testing.T) {
	p, client, cursor := generatTenorProviderForURLBuildingTests()
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=off")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterLanguageEmpty(t *testing.T) {
	p, client, cursor := generatTenorProviderForURLBuildingTests()
	p.language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "locale")
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterLanguageProvided(t *testing.T) {
	p, client, cursor := generatTenorProviderForURLBuildingTests()
	p.language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "locale="+p.language)
		return true
	}
	_, err := p.GetGifURL("cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
