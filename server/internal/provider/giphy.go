package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v5/model"
)

// Giphy find GIFs using the Giphy API
type Giphy struct {
	abstractGifProvider
}

const (
	baseURLGiphy = "https://api.giphy.com/v1/gifs"
)

type giphySearchResult struct {
	Data []struct {
		Images map[string]struct {
			URL string `json:"url"`
		} `json:"images"`
	} `json:"data"`
	Pagination struct {
		Offset int `json:"offset"`
	} `json:"pagination"`
}

// NewGiphyProvider creates an instance of a GIF provider that uses the Giphy API
func NewGiphyProvider(httpClient HTTPClient, errorGenerator pluginError.PluginError) GifProvider {
	giphyProvider := &Giphy{}
	giphyProvider.httpClient = httpClient
	giphyProvider.errorGenerator = errorGenerator
	return giphyProvider
}

func (p *Giphy) GetAttributionMessage() string {
	return "Powered by Giphy"
}

func (p *Giphy) GetGifURL(config *pluginConf.Configuration, request string, cursor *string) (string, *model.AppError) {
	if config.APIKey == "" {
		return "", p.errorGenerator.FromMessage("Giphy API key is empty")
	}
	req, err := http.NewRequest("GET", baseURLGiphy+"/search", nil)
	if err != nil {
		return "", p.errorGenerator.FromError("Could not generate URL", err)
	}

	q := req.URL.Query()

	q.Add("api_key", config.APIKey)
	q.Add("q", request)
	if counter, err2 := strconv.Atoi(*cursor); err2 == nil {
		q.Add("offset", fmt.Sprintf("%d", counter))
	}
	q.Add("limit", "1")
	if len(config.Rating) > 0 {
		q.Add("rating", config.Rating)
	}
	if len(config.Language) > 0 {
		q.Add("lang", config.Language)
	}

	req.URL.RawQuery = q.Encode()

	r, err := p.httpClient.Do(req)
	if err != nil {
		return "", p.errorGenerator.FromError("Error calling the Giphy API", err)
	}

	if r.StatusCode != http.StatusOK {
		explanation := ""
		if r.StatusCode == http.StatusTooManyRequests {
			explanation = ", this can happen if you're using the default GIPHY API key"
		}
		return "", p.errorGenerator.FromMessage(fmt.Sprintf("Error calling the Giphy API (HTTP Status: %v%s)", r.Status, explanation))
	}
	var response giphySearchResult
	if r.Body == nil {
		return "", p.errorGenerator.FromMessage("GIPHY search response body is empty")
	}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", p.errorGenerator.FromError("Could not parse GIPHY search response body", err)
	}
	if len(response.Data) < 1 {
		return "", p.errorGenerator.FromMessage("No more GIF results for this search!")
	}
	gif := response.Data[0]
	url := gif.Images[config.Rendition].URL

	if len(url) < 1 {
		return "", p.errorGenerator.FromMessage("No URL found for display style \"" + config.Rendition + "\" in the response")
	}
	*cursor = fmt.Sprintf("%d", response.Pagination.Offset+1)
	return url, nil
}
