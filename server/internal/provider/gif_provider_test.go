package provider

import (
	"reflect"
	"testing"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestDefaultGifProviderGenerator(t *testing.T) {
	testCases := []struct {
		testLabel     string
		providerType  string
		expectedError bool
		expectedType  interface{}
	}{
		{testLabel: "Empty provider", providerType: "", expectedError: true, expectedType: nil},
		{testLabel: "Giphyprovider", providerType: "giphy", expectedError: false, expectedType: &giphy{}},
		{testLabel: "Tenor provider", providerType: "tenor", expectedError: false, expectedType: tenor{}},
	}

	for _, testCase := range testCases {
		testConfig := pluginConf.Configuration{Provider: testCase.providerType,
			APIKey:         testGiphyAPIKey,
			Language:       testGiphyLanguage,
			Rating:         testGiphyRating,
			Rendition:      testGiphyRendition,
			RenditionTenor: testTenorRendition,
		}
		provider, err := defaultGifProviderGenerator(testConfig, test.MockErrorGenerator(), "/test")
		if testCase.expectedError {
			assert.NotNil(t, err, testCase.testLabel)
			assert.Nil(t, provider, testCase.testLabel)
		} else {
			assert.Nil(t, err, testCase.testLabel)
			assert.NotNil(t, provider, testCase.testLabel)
			assert.Equal(t, "*provider."+testConfig.Provider, reflect.TypeOf(provider).String())
		}
	}
}
