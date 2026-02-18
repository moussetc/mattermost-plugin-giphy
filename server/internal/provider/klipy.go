package provider

import (
	"encoding/json"
	"fmt"
	"net/http"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost/server/public/model"
)

// NewKlipyProvider creates an instance of a GIF provider that uses the KLIPY API
func NewKlipyProvider(httpClient HTTPClient, errorGenerator pluginError.PluginError, apiKey, language, rating, rendition string, rootURL string) (GifProvider, *model.AppError) {
	if errorGenerator == nil {
		return nil, model.NewAppError("NewKlipyProvider", "errorGenerator cannot be nil for KLIPY Provider", nil, "", http.StatusInternalServerError)
	}
	if httpClient == nil {
		return nil, errorGenerator.FromMessage("httpClient cannot be nil for KLIPY Provider")
	}
	if apiKey == "" {
		return nil, errorGenerator.FromMessage("apiKey cannot be empty for KLIPY Provider")
	}
	if rendition == "" {
		return nil, errorGenerator.FromMessage("rendition cannot be empty for KLIPY Provider")
	}

	klipyProvider := klipy{}
	klipyProvider.rootURL = rootURL
	klipyProvider.httpClient = httpClient
	klipyProvider.errorGenerator = errorGenerator
	klipyProvider.apiKey = apiKey
	klipyProvider.language = language
	klipyProvider.rating = convertKlipyRatingToContentFilter(rating)
	klipyProvider.rendition = rendition

	return &klipyProvider, nil
}

// klipy finds GIFs using the KLIPY API (Tenor-v2 compatible base URL swap)
type klipy struct {
	abstractGifProvider
	apiKey  string
	rootURL string
}

const (
	baseURLKlipy = "https://api.klipy.com/v2" // Tenor migration: tenor.googleapis.com -> api.klipy.com
)

type klipySearchResult struct {
	Next    string `json:"next"`
	Results []struct {
		Media map[string]struct {
			URL string `json:"url"`
		} `json:"media_formats"`
	} `json:"results"`
}

type klipySearchError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (p *klipy) GetAttributionMessage() string {
	return fmt.Sprintf("![KLIPY](%s/public/powered-by-klipy.png)", p.rootURL)
}

// Return URLs of GIFs that match the query, or empty slice if none, or an error if the search failed
func (p *klipy) GetGifURL(request string, cursor *string, random bool) ([]string, *model.AppError) {
	req, err := http.NewRequest("GET", baseURLKlipy+"/search", nil)
	if err != nil {
		return []string{}, p.errorGenerator.FromError("Could not generate URL", err)
	}

	q := req.URL.Query()

	q.Add("key", p.apiKey)
	q.Add("q", request)
	q.Add("ar_range", "all")
	if cursor != nil && *cursor != "" {
		q.Add("pos", *cursor)
	}

	// if random, we need to have several results because random=true applies to the result list of this query
	q.Add("contentfilter", p.rating)
	q.Add("media_filter", p.rendition)
	if len(p.language) > 0 {
		q.Add("locale", p.language)
	}
	if random {
		q.Add("random", "true")
	}

	req.URL.RawQuery = q.Encode()

	r, err := p.httpClient.Do(req)
	if err != nil {
		return []string{}, p.errorGenerator.FromError("Error calling the KLIPY API", err)
	}
	if r != nil && r.Body != nil {
		defer r.Body.Close()
	}

	if r.StatusCode != http.StatusOK {
		var errorResponse klipySearchError
		errorDetails := fmt.Sprintf("Error calling the KLIPY API (HTTP Status: %v", r.Status)
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			if err = decoder.Decode(&errorResponse); err == nil && errorResponse.Error != "" {
				errorDetails += ", API message: \"" + errorResponse.Error + "\""
			}
		}
		errorDetails += ")"
		return []string{}, p.errorGenerator.FromMessage(errorDetails)
	}

	var response klipySearchResult
	if r.Body == nil {
		return []string{}, p.errorGenerator.FromMessage("KLIPY search response body is empty")
	}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return []string{}, p.errorGenerator.FromError("Could not parse KLIPY search response body", err)
	}

	if len(response.Results) < 1 {
		return []string{}, nil
	}

	urls := []string{}
	for i := range response.Results {
		url := response.Results[i].Media[p.rendition].URL
		if len(url) > 0 {
			urls = append(urls, url)
		}
	}

	if len(urls) < 1 {
		return []string{}, p.errorGenerator.FromMessage("No gifs found for display style \"" + p.rendition + "\" in the response")
	}

	if cursor != nil {
		*cursor = response.Next
	}

	return urls, nil
}

func convertKlipyRatingToContentFilter(rating string) string {
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
