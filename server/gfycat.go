package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
	"net/http"
	"strings"
)

// gifyCatProvider get GIF URLs from the GfyCat API, using Mattermost settings
type gfyCatProvider struct{}

const (
	GFYCAT_BASE_URL = "https://api.gfycat.com/v1"
)

type gfySearchResult struct {
	Cursor  string                        `json:"cursor"`
	Gfycats []map[string]*json.RawMessage `json:"gfycats"`
}

func (p *gfyCatProvider) getAttributionMessage() string {
	return "Powered by Gfycat"
}

// getGifURL return the URL of a GIF that matches the requested keywords
func (p *gfyCatProvider) getGifURL(config *configuration, request string, cursor *string) (string, *model.AppError) {
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
		return "", appError(fmt.Sprintf("Error calling the GfyCat search API (HTTP Status: %v)", r.Status), nil)
	}
	var response gfySearchResult
	decoder := json.NewDecoder(r.Body)
	if r.Body == nil {
		return "", appError("GfyCat search response body is empty", nil)
	}
	if err = decoder.Decode(&response); err != nil {
		return "", appError("Could not parse Gfycat search response body", err)
	}
	if len(response.Gfycats) < 1 {
		return "", appError("No more GIF results for this search!", nil)
	}
	gif := response.Gfycats[0]
	urlNode, ok := gif[(*config).RenditionGfycat]
	if !ok {
		return "", appError("No URL found for display style \""+(*config).RenditionGfycat+"\" in the response", nil)
	}
	var url string
	if urlNode != nil {
		if err = json.Unmarshal(*urlNode, &url); err != nil {
			return "", appError("Could not read "+(*config).RenditionGfycat+"node", err)
		}
	}
	// Ignore suffix without a Mattermost preview
	if url == "" || strings.HasSuffix(url, ".webm") || strings.HasSuffix(url, ".mp4") {
		urlNode, ok = gif["gifUrl"]
		if !ok {
			return "", appError("No URL found for the \"gifUrl\" in the response", nil)
		}
		if err = json.Unmarshal(*urlNode, &url); err != nil {
			return "", appError("Could not read gifUrl node", err)
		}
	}
	if url == "" {
		return "", appError("An empty URL was returned for display style \""+config.RenditionGfycat+"\"", nil)
	}
	*cursor = response.Cursor
	return url, nil
}
