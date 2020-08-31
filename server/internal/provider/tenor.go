package provider

import (
	"encoding/json"
	"fmt"
	"net/http"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v5/model"
)

// NewTenorProvider creates an instance of a GIF provider that uses the Tenor API
func NewTenorProvider(httpClient HTTPClient, errorGenerator pluginError.PluginError, apiKey, language, rating, rendition string) (GifProvider, *model.AppError) {
	if apiKey == "" {
		return nil, errorGenerator.FromMessage("The API Key setting must be set for Tenor")
	}
	if rendition == "" {
		return nil, errorGenerator.FromMessage("The Rendition setting must be set for Tenor")
	}

	tenorProvider := Tenor{}
	tenorProvider.httpClient = httpClient
	tenorProvider.errorGenerator = errorGenerator
	tenorProvider.apiKey = apiKey
	tenorProvider.language = language
	tenorProvider.rating = convertRatingToContentFilter(rating)
	tenorProvider.rendition = rendition

	return &tenorProvider, nil
}

// Tenor find GIFs using the Tenor API
type Tenor struct {
	abstractGifProvider
	apiKey string
}

const (
	baseURLTenor = "https://api.tenor.com/v1"
)

type tenorSearchResult struct {
	Next    string `json:"next"`
	Results []struct {
		Media []map[string]struct {
			URL string `json:"url"`
		} `json:"media"`
	} `json:"results"`
}

type tenorSearchError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (p *Tenor) GetAttributionMessage() string {
	return "Via Tenor"
}

// Return the URL of a GIF that matches the requested keywords
func (p *Tenor) GetGifURL(request string, cursor *string) (string, *model.AppError) {
	req, err := http.NewRequest("GET", baseURLTenor+"/search", nil)
	if err != nil {
		return "", p.errorGenerator.FromError("Could not generate URL", err)
	}

	q := req.URL.Query()

	q.Add("key", p.apiKey)
	q.Add("q", request)
	q.Add("ar_range", "all")
	if cursor != nil && *cursor != "" {
		q.Add("pos", *cursor)
	}
	q.Add("limit", "1")
	q.Add("contentfilter", p.rating)
	if len(p.language) > 0 {
		q.Add("locale", p.language)
	}

	req.URL.RawQuery = q.Encode()

	r, err := p.httpClient.Do(req)
	if err != nil {
		return "", p.errorGenerator.FromError("Error calling the Tenor API", err)
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
		return "", p.errorGenerator.FromMessage(errorDetails)
	}
	var response tenorSearchResult
	if r.Body == nil {
		return "", p.errorGenerator.FromMessage("Tenor search response body is empty")
	}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", p.errorGenerator.FromError("Could not parse Tenor search response body", err)
	}
	if len(response.Results) < 1 || len(response.Results[0].Media) < 1 {
		return "", p.errorGenerator.FromMessage("No more GIF results for this search!")
	}
	gif := response.Results[0].Media[0]
	url := gif[p.rendition].URL

	if len(url) < 1 {
		return "", p.errorGenerator.FromMessage("No URL found for display style \"" + p.rendition + "\" in the response")
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
