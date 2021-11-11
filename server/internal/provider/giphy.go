package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v5/model"
)

// giphy find GIFs using the giphy API
type giphy struct {
	abstractGifProvider
	apiKey  string
	rootURL string
}

const (
	baseURLGiphy = "https://api.Giphy.com/v1/gifs"
)

type GiphySearchResult struct {
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
func NewGiphyProvider(httpClient HTTPClient, errorGenerator pluginError.PluginError, apiKey, language, rating, rendition, rootURL string) (GifProvider, *model.AppError) {
	if errorGenerator == nil {
		return nil, model.NewAppError("NewGfycatProvider", "errorGenerator cannot be nil for Giphy Provider", nil, "", http.StatusInternalServerError)
	}
	if httpClient == nil {
		return nil, errorGenerator.FromMessage("httpClient cannot be nil for Giphy Provider")
	}
	if apiKey == "" {
		return nil, errorGenerator.FromMessage("apiKey cannot be empty for Giphy Provider")
	}
	if rendition == "" {
		return nil, errorGenerator.FromMessage("rendition cannot be empty for Giphy Provider")
	}
	if rootURL == "" {
		return nil, errorGenerator.FromMessage("internal error: rootURL must be set")
	}

	GiphyProvider := &giphy{}
	GiphyProvider.httpClient = httpClient
	GiphyProvider.errorGenerator = errorGenerator
	GiphyProvider.apiKey = apiKey
	GiphyProvider.language = language
	GiphyProvider.rating = rating
	GiphyProvider.rendition = rendition
	GiphyProvider.rootURL = rootURL

	return GiphyProvider, nil
}

func (p *giphy) GetAttributionMessage() string {
	return fmt.Sprintf("![GIPHY](%s/public/powered-by-giphy.png)", p.rootURL)
}

// Return the URL of a GIF that matches the query, or an empty string if no GIF matches the query, or an error if the search failed
func (p *giphy) GetGifURL(request string, cursor *string) (string, *model.AppError) {
	req, err := http.NewRequest("GET", baseURLGiphy+"/search", nil)
	if err != nil {
		return "", p.errorGenerator.FromError("Could not generate URL", err)
	}

	q := req.URL.Query()

	q.Add("api_key", p.apiKey)
	q.Add("q", request)
	if counter, err2 := strconv.Atoi(*cursor); err2 == nil {
		q.Add("offset", fmt.Sprintf("%d", counter))
	}
	q.Add("limit", "1")
	if len(p.rating) > 0 {
		q.Add("rating", p.rating)
	}
	if len(p.language) > 0 {
		q.Add("lang", p.language)
	}

	req.URL.RawQuery = q.Encode()

	r, err := p.httpClient.Do(req)
	if err != nil {
		return "", p.errorGenerator.FromError("Error calling the Giphy API", err)
	}

	if r.StatusCode != http.StatusOK {
		explanation := ""
		if r.StatusCode == http.StatusTooManyRequests {
			explanation = ", this can happen if you're using the default Giphy API key"
		}
		return "", p.errorGenerator.FromMessage(fmt.Sprintf("Error calling the Giphy API (HTTP Status: %v%s)", r.Status, explanation))
	}
	var response GiphySearchResult
	if r.Body == nil {
		return "", p.errorGenerator.FromMessage("Giphy search response body is empty")
	}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", p.errorGenerator.FromError("Could not parse Giphy search response body", err)
	}
	if len(response.Data) < 1 {
		return "", nil
	}
	gif := response.Data[0]
	url := gif.Images[p.rendition].URL

	if len(url) < 1 {
		return "", p.errorGenerator.FromMessage("No URL found for display style \"" + p.rendition + "\" in the response")
	}
	*cursor = fmt.Sprintf("%d", response.Pagination.Offset+1)
	return url, nil
}
