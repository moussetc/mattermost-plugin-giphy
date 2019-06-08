package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// gifyCatProvider get GIF URLs from the GfyCat API, using Mattermost settings
type gfyCatProvider struct{}

const (
	GFYCAT_BASE_URL = "https://api.gfycat.com/v1"
)

type gfySearchResult struct {
	Cursor  string   `json:"cursor"`
	Gfycats []gfyGIF `json:"gfycats"`
}

type gfyGIF struct {
	GifUrl      string `json:"gifUrl"`
	ContentUrls map[string]struct {
		Url string `json:"url"`
	} `json:"content_urls"`
}

// getGifURL return the URL of a GIF that matches the requested keywords
func (p *gfyCatProvider) getGifURL(config *configuration, request string, cursor *string) (string, error) {
	req, err := http.NewRequest("GET", GFYCAT_BASE_URL+"/gfycats/search", nil)
	if err != nil {
		return "", appError("Could not generate GfyCat search URL", err)
	}
	q := req.URL.Query()
	q.Add("search_text", request)
	q.Add("count", "1")
	if *cursor != "" {
		q.Add("cursor", *cursor)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	r, err := getGifProviderHttpClient().Do(req)
	if err != nil {
		return "", appError("Error calling the GfyCat search API", err)
	}

	if r.StatusCode != http.StatusOK {
		return "", appError(fmt.Sprintf("Error calling the GfyCat search API (HTTP Status: %v)", r.StatusCode), nil)
	}
	var response gfySearchResult
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", appError("Could not parse Gfycat search response body", err)
	}
	if len(response.Gfycats) < 1 {
		return "", appError("An empty list of GIFs was returned", err)
	}
	gif := response.Gfycats[0]
	url := gif.ContentUrls[(*config).RenditionGfycat].Url
	// Ignore suffix without a Mattermost preview
	if url == "" || strings.HasSuffix(url, ".webm") || strings.HasSuffix(url, ".mp4") {
		url = gif.GifUrl
	}
	*cursor = response.Cursor
	return url, nil
}
