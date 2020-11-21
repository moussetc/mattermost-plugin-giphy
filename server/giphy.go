package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
)

// giphyProvider get GIF URLs from the Giphy API without any external, out-of-date library
type giphyProvider struct{}

const (
	BASE_URL = "https://api.giphy.com/v1/gifs"
)

type giphySearchResult struct {
	Data []struct {
		Images map[string]struct {
			Url string `json:"url"`
		} `json:"images"`
	} `json:"data"`
	Pagination struct {
		Offset int `json:"offset"`
	} `json:"pagination"`
}

func (p *giphyProvider) getAttributionMessage() string {
	return fmt.Sprintf("![GIPHY](/plugins/%s/public/powered-by-giphy.png)", manifest.Id)
}

// getGifURL return the URL of a GIF that matches the requested keywords
func (p *giphyProvider) getGifURL(config *configuration, request string, cursor *string) (string, *model.AppError) {
	if config.APIKey == "" {
		return "", appError("Giphy API key is empty", nil)
	}
	req, err := http.NewRequest("GET", BASE_URL+"/search", nil)
	if err != nil {
		return "", appError("Could not generate URL", err)
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

	r, err := getGifProviderHttpClient().Do(req)
	if err != nil {
		return "", appError("Error calling the Giphy API", err)
	}

	if r.StatusCode != http.StatusOK {
		explanation := ""
		if r.StatusCode == http.StatusTooManyRequests {
			explanation = ", this can happen if you're using the default GIPHY API key"
		}
		return "", appError(fmt.Sprintf("Error calling the Giphy API (HTTP Status: %v%s)", r.Status, explanation), nil)
	}
	var response giphySearchResult
	if r.Body == nil {
		return "", appError("GIPHY search response body is empty", nil)
	}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", appError("Could not parse GIPHY search response body", err)
	}
	if len(response.Data) < 1 {
		return "", appError("No more GIF results for this search!", nil)
	}
	gif := response.Data[0]
	url := gif.Images[config.Rendition].Url

	if len(url) < 1 {
		return "", appError("No URL found for display style \""+config.Rendition+"\" in the response", nil)
	}
	*cursor = fmt.Sprintf("%d", response.Pagination.Offset+1)
	return url, nil
}
