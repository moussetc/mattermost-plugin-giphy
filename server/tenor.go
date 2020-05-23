package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

type tenorProvider struct{}

const (
	baseURL = "https://api.tenor.com/v1"
)

type tenorSearchResult struct {
	Next    string `json:"next"`
	Results []struct {
		Media []map[string]struct {
			Url string `json:"url"`
		} `json:"media"`
	} `json:"results"`
}

type tenorSearchError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Return the URL of a GIF that matches the requested keywords
func (p *tenorProvider) getGifURL(config *configuration, request string, cursor *string) (string, *model.AppError) {
	if config.APIKey == "" {
		return "", appError("Tenor API key is empty", nil)
	}
	req, err := http.NewRequest("GET", baseURL+"/search", nil)
	if err != nil {
		return "", appError("Could not generate URL", err)
	}

	q := req.URL.Query()

	q.Add("key", config.APIKey)
	q.Add("q", request)
	q.Add("ar_range", "all")
	if cursor != nil && *cursor != "" {
		q.Add("pos", *cursor)
	}
	q.Add("limit", "1")
	q.Add("contentfilter", convertRatingToContentFilter(config.Rating))
	if len(config.Language) > 0 {
		q.Add("locale", config.Language)
	}

	req.URL.RawQuery = q.Encode()

	r, err := getGifProviderHttpClient().Do(req)
	if err != nil {
		return "", appError("Error calling the Tenor API", err)
	}

	if r.StatusCode != http.StatusOK {
		var errorResponse tenorSearchError
		errorDetails := fmt.Sprintf("Error calling the Tenor API (HTTP Status: %v", r.Status)
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			if err = decoder.Decode(&errorResponse); err == nil && errorResponse.Error != "" {
				errorDetails += ", API message: \"" + errorResponse.Error + "\""
			}
		}
		errorDetails += ")"
		return "", appError(errorDetails, nil)
	}
	var response tenorSearchResult
	if r.Body == nil {
		return "", appError("Tenor search response body is empty", nil)
	}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", appError("Could not parse Tenor search response body", err)
	}
	if len(response.Results) < 1 || len(response.Results[0].Media) < 1 {
		return "", appError("No more GIF results for this search!", nil)
	}
	gif := response.Results[0].Media[0]
	url := gif[config.RenditionTenor].Url

	if len(url) < 1 {
		return "", appError("No URL found for display style \""+config.RenditionTenor+"\" in the response", nil)
	}
	*cursor = response.Next
	return url, nil
}

func convertRatingToContentFilter(rating string) string {
	switch rating {
	case "g":
		return "high"
	case "pg":
		return "medium"
	case "pg-13":
		return "low"
	default:
		return "off"
	}
}
