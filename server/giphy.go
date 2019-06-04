package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// giphyProvider get GIF URLs from the Giphy API without any external, out-of-date library
type giphyProvider struct{}

const (
	BASE_URL = "http://api.giphy.com/v1/gifs"
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

// getGifURL return the URL of a GIF that matches the requested keywords
func (p *giphyProvider) getGifURL(config *configuration, request string, cursor *string) (string, error) {
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
		q.Add("offset", fmt.Sprintf("%d", counter+1))
	}
	q.Add("limit", "1")
	if len(config.Rating) > 0 {
		q.Add("rating", config.Rating)
	}
	if len(config.Language) > 0 {
		q.Add("lang", config.Language)
	}

	req.URL.RawQuery = q.Encode()
	requestURL := req.URL.String()

	r, err := http.DefaultClient.Get(requestURL)
	if err != nil {
		return "", appError("Error calling the Giphy API", err)
	}

	if r.StatusCode != http.StatusOK {
		return "", appError("Error calling the Giphy API (HTTP Status: "+fmt.Sprintf("%v", r.StatusCode), nil)
	}
	var response giphySearchResult
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", appError("Could not parse GIPHY search response body", err)
	}
	if len(response.Data) < 1 {
		return "", appError("An empty list of GIFs was returned", nil)
	}
	gif := response.Data[0]
	url := gif.Images[config.Rendition].Url

	*cursor = fmt.Sprintf("%d", response.Pagination.Offset)
	return url, nil
}
