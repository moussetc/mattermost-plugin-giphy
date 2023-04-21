package provider

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v6/model"
)

// NewGfycatProvider creates an instance of a GIF provider that uses the GfyCat API
func NewGfycatProvider(httpClient HTTPClient, errorGenerator pluginError.PluginError, rendition string) (GifProvider, *model.AppError) {
	if errorGenerator == nil {
		return nil, model.NewAppError("NewGfycatProvider", "errorGenerator cannot be nil for Gfycat Provider", nil, "", http.StatusInternalServerError)
	}
	if httpClient == nil {
		return nil, errorGenerator.FromMessage("httpClient cannot be nil for Gfycat Provider")
	}
	if rendition == "" {
		return nil, errorGenerator.FromMessage("rendition cannot be empty for Gfycat Provider")
	}

	gfycatProvider := gfycat{}
	gfycatProvider.httpClient = httpClient
	gfycatProvider.errorGenerator = errorGenerator
	gfycatProvider.rendition = rendition

	return &gfycatProvider, nil
}

// gfycat find GIFs using the GfyCat API
type gfycat struct {
	abstractGifProvider
}

const (
	baseURLGfycat = "https://api.gfycat.com/v1"
)

type gfySearchResult struct {
	Cursor  string                        `json:"cursor"`
	Gfycats []map[string]*json.RawMessage `json:"gfycats"`
}

func (p *gfycat) GetAttributionMessage() string {
	return "Powered by Gfycat"
}

// Return the URL of a GIF that matches the query, or an empty string if no GIF matches the query, or an error if the search failed
func (p *gfycat) GetGifURL(request string, cursor *string, random bool) ([]string, *model.AppError) {
	/**
	 * Known quirks of the Gfycat API
	 * - "count" parameter is applied _before_ any filtering (private GIF, etc.) so if you ask
	 *   for N you only know will get a "page" of *at most* N, regardless of if GIFs remains
	 *   in the next "pages". Source:
	 *   https://www.reddit.com/r/gfycat/comments/ijjq5n/api_why_a_request_with_count1_returns_0_results/
	 *
	 * - the cursor property applies to the whole "page"
	 *
	 * => instead of using "count=1" (and get possibly empty result if the cursor points to a private GIF), we
	 * get a whole page of GIFs and iterate manually a cursor within this page.
	**/
	req, err := http.NewRequest("GET", baseURLGfycat+"/gfycats/search", nil)
	if err != nil {
		return []string{}, p.errorGenerator.FromError("Could not generate GfyCat search URL", err)
	}

	q := req.URL.Query()

	q.Add("search_text", request)
	if cursor != nil && *cursor != "" {
		q.Add("cursor", *cursor)
	}

	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	r, err := p.httpClient.Do(req)
	if err != nil {
		return []string{}, p.errorGenerator.FromError("Error calling the GfyCat search API", err)
	}
	if r != nil && r.Body != nil {
		defer r.Body.Close()
	}

	if r.StatusCode != http.StatusOK {
		return []string{}, p.errorGenerator.FromMessage(fmt.Sprintf("Error calling the GfyCat search API (HTTP Status: %v)", r.Status))
	}

	var response gfySearchResult
	decoder := json.NewDecoder(r.Body)
	if r.Body == nil {
		return []string{}, p.errorGenerator.FromMessage("GfyCat search response body is empty")
	}
	if err = decoder.Decode(&response); err != nil {
		return []string{}, p.errorGenerator.FromError("Could not parse Gfycat search response body", err)
	}
	if len(response.Gfycats) < 1 {
		return []string{}, nil
	}

	urls := []string{}
	for i := range response.Gfycats {
		gif := response.Gfycats[i]
		urlNode, ok := gif[p.rendition]
		if ok && urlNode != nil {
			var url string
			if err = json.Unmarshal(*urlNode, &url); err == nil {
				// Ignore suffix without a Mattermost preview
				if url == "" || strings.HasSuffix(url, ".webm") || strings.HasSuffix(url, ".mp4") {
					// Use gifUrl as fallback
					if urlNode, ok = gif["gifUrl"]; ok {
						if err = json.Unmarshal(*urlNode, &url); err == nil {
							urls = append(urls, url)
						}
					}
				} else {
					urls = append(urls, url)
				}
			}
		}
	}

	if len(urls) < 1 {
		return []string{}, p.errorGenerator.FromMessage("No gifs found for display style \"" + p.rendition + "\" in the response")
	}

	*cursor = response.Cursor

	// As the API does not provide a random endpoint of option, we return a randomized page as best effort
	if random {
		for i := len(urls) - 1; i > 0; i-- {
			j := rand.Intn(i + 1) //nolint:gosec
			urls[i], urls[j] = urls[j], urls[i]
		}
	}

	return urls, nil
}
