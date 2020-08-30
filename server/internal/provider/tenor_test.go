package provider

import (
	"net/http"
	"testing"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

const defaultTenorResponseBody = "{  \"weburl\": \"https://fakeurl/search/stuff\",   \"results\": [   {    \"tags\": [],     \"url\": \"https://fakeurl/fake.gif\",     \"media\": [     {      \"nanomp4\": {       \"url\": \"https://fakeurl/nanomp4\",        \"dims\": [        150,         112       ],        \"duration\": 0.5,        \"preview\": \"https://fakeurl/preview.png\",        \"size\": 7343      },       \"nanowebm\": {       \"url\": \"https://fakeurl/nanowebm\",        \"dims\": [        150,         112       ],        \"preview\": \"https://fakeurl/nanowebm/preview.png\",        \"size\": 9550      },       \"tinygif\": {       \"url\": \"https://fakeurl/tinygif.gif\",        \"dims\": [        220,         164       ],        \"preview\": \"https://fakeurl/tinigif/preview.gif\",        \"size\": 22519      },       \"tinymp4\": {       \"url\": \"https://fakeurl/mp4\",        \"dims\": [        320,         238       ],        \"duration\": 0.5,        \"preview\": \"https://fakeurl/tinymp4/preview.png\",        \"size\": 17732      },       \"tinywebm\": {       \"url\": \"https://fakeurl/tinywebm\",        \"dims\": [        320,         238       ],        \"preview\": \"https://fakeurl/tinywebm/preview.png\",        \"size\": 12311      },       \"webm\": {       \"url\": \"https://fakeurl/webm\",        \"dims\": [        444,         332       ],        \"preview\": \"https://fakeurl/webm/preview.png\",        \"size\": 14924      },       \"gif\": {       \"url\": \"https://fakeurl/gif.gif\",        \"dims\": [        444,         332       ],        \"preview\": \"https://fakeurl/gif/preview.png\",        \"size\": 465547      },       \"mp4\": {       \"url\": \"https://fakeurl/mp4\",        \"dims\": [        444,         332       ],        \"duration\": 0.5,        \"preview\": \"https://fakeurl/mp4/preview.png\",        \"size\": 36818      },       \"loopedmp4\": {       \"url\": \"https://fakeurl/loopedmp4\",        \"dims\": [        444,         332       ],        \"duration\": 1.5,        \"preview\": \"https://fakeurl/loopedmp4/preview.png\",        \"size\": 108909      },       \"mediumgif\": {       \"url\": \"https://fakeurl/mediumgif.gif\",        \"dims\": [        444,         332       ],        \"preview\": \"https://fakeurl/imediumgif/preview.gif\",        \"size\": 127524      },       \"nanogif\": {       \"url\": \"https://fakeurl/nanogif.gif\",        \"dims\": [        120,         90       ],        \"preview\": \"https://fakeurl/nanogif/preview.gif\",        \"size\": 9104      }     }    ],     \"created\": 1476975012.524378,     \"shares\": 1,     \"itemurl\": \"https://fakeurl/view/fakeurl\",     \"composite\": null,     \"hasaudio\": false,     \"title\": \"\",     \"id\": \"6198981\"   }  ],   \"next\": \"1\" }"

func generateMockConfigForTenorProvider() pluginConf.Configuration {
	return pluginConf.Configuration{
		APIKey:         "defaultAPIKey",
		Rating:         "",
		Language:       "fr",
		RenditionTenor: "mediumgif",
	}
}

func TestTenorProviderGetGifURLOK(t *testing.T) {
	p := NewTenorProvider(NewMockHttpClient(newServerResponseOK(defaultTenorResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	config.RenditionTenor = "tinygif"
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
	assert.Equal(t, url, "https://fakeurl/tinygif.gif")
}

func TestTenorProviderMissingAPIKey(t *testing.T) {
	p := NewTenorProvider(NewMockHttpClient(newServerResponseOK(defaultTenorResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	config.APIKey = ""
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "API key")
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLEmptyBody(t *testing.T) {
	p := NewTenorProvider(NewMockHttpClient(newServerResponseOK("")), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.Empty(t, url)
}

func TestTenorProviderGetGifURLParseError(t *testing.T) {
	p := NewTenorProvider(NewMockHttpClient(newServerResponseOK("This is not a valid JSON response")), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Empty(t, url)
}

func TestTenorProviderEmptyGIFList(t *testing.T) {
	p := NewTenorProvider(NewMockHttpClient(newServerResponseOK("{ \"weburl\": \"https://fakeurl/casdfsdfsdfsdfsdfst-gifs\", \"results\": [], \"next\": \"0\" }")), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No more GIF result")
	assert.Empty(t, url)
}

func TestTenorProviderEmptyURLForRendition(t *testing.T) {
	p := NewTenorProvider(NewMockHttpClient(newServerResponseOK(defaultTenorResponseBody)), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	config.RenditionTenor = "NotExistingDisplayStyle"
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "No URL found for display style")
	assert.Contains(t, err.Error(), config.RenditionTenor)
	assert.Empty(t, url)
}

func TestTenorProviderErrorStatusResponseWithoutErrorMessage(t *testing.T) {
	serverResponse := newServerResponseKO(400)
	p := NewTenorProvider(NewMockHttpClient(serverResponse), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Empty(t, url)
}

func TestTenorProviderErrorStatusResponseWithErrorMessage(t *testing.T) {
	serverResponse := newServerResponseKOWithBody(429, "{ \"error\": \"Please use a registered API Key\" }")
	p := NewTenorProvider(NewMockHttpClient(serverResponse), test.MockErrorGenerator())
	config := generateMockConfigForTenorProvider()
	cursor := ""
	url, err := p.GetGifURL(&config, "cat", &cursor)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), serverResponse.Status)
	assert.Contains(t, err.Error(), "Please use a registered API Key")
	assert.Empty(t, url)
}

func generateHTTPClientForTenorParameterTest() (p GifProvider, client *MockHttpClient, config pluginConf.Configuration, cursor string) {
	serverResponse := newServerResponseOK(defaultTenorResponseBody)
	client = NewMockHttpClient(serverResponse)
	p = NewTenorProvider(client, test.MockErrorGenerator())
	config = generateMockConfigForTenorProvider()
	cursor = ""
	return p, client, config, cursor
}

func TestTenorProviderParameterRatingProvidedG(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForTenorParameterTest()

	config.Rating = "g"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=high")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterRatingProvidedPG(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForTenorParameterTest()

	config.Rating = "pg"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=medium")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterRatingProvidedPG13(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForTenorParameterTest()

	config.Rating = "pg-13"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=low")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterRatingProvidedR(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForTenorParameterTest()

	config.Rating = "r"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=off")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterRatingNotProvided(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForTenorParameterTest()

	config.Rating = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "contentfilter=off")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterLanguageEmpty(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForTenorParameterTest()

	config.Language = ""
	client.testRequestFunc = func(req *http.Request) bool {
		assert.NotContains(t, req.URL.RawQuery, "locale")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}

func TestTenorProviderParameterLanguageProvided(t *testing.T) {
	p, client, config, cursor := generateHTTPClientForTenorParameterTest()

	// Initial value : 0
	config.Language = "Moldovalaque"
	client.testRequestFunc = func(req *http.Request) bool {
		assert.Contains(t, req.URL.RawQuery, "locale=Moldovalaque")
		return true
	}
	_, err := p.GetGifURL(&config, "cat", &cursor)
	assert.Nil(t, err)
	assert.True(t, client.lastRequestPassTest)
}
