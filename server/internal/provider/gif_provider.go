package provider

import (
	"net/http"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v5/model"
)

// GifProvider exposes methods to get GIF from an API
type GifProvider interface {
	// GetGifURL return the URL of a GIF that matches the requested keywords if one is found or else
	GetGifURL(request string, cursor *string) (string, *model.AppError)

	// GetAttributionMessage returns the text that should be displayed near the GIF, as defined by the providers' Terms of Service
	GetAttributionMessage() string
}

// HTTPClient is an subset of the standard HTTP client functions used by GIF Providers
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(s string) (*http.Response, error)
}

type Query struct {
	Keywords string
	Cursor   string
}

type abstractGifProvider struct {
	httpClient     HTTPClient
	errorGenerator pluginError.PluginError
	language       string
	rating         string
	rendition      string
}

func defaultGifProviderGenerator(configuration pluginConf.Configuration, errorGenerator pluginError.PluginError, rootURL string) (gifProvider GifProvider, err *model.AppError) {
	if configuration.Provider == "" {
		return nil, errorGenerator.FromMessage("The GIF provider must be configured")
	}
	switch configuration.Provider {
	case "giphy":
		gifProvider, err = NewGiphyProvider(http.DefaultClient, errorGenerator, configuration.APIKey, configuration.Language, configuration.Rating, configuration.Rendition, rootURL)
	case "tenor":
		gifProvider, err = NewTenorProvider(http.DefaultClient, errorGenerator, configuration.APIKey, configuration.Language, configuration.Rating, configuration.RenditionTenor)
	default:
		gifProvider, err = NewGfycatProvider(http.DefaultClient, errorGenerator, configuration.RenditionGfycat)
	}
	return gifProvider, err
}

var GifProviderGenerator = defaultGifProviderGenerator
