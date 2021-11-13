package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v5/model"
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

type gfyPageCursor struct {
	CursorForPage  string `json:"cursorForPage"`
	PositionInPage int    `json:"positionInPage"`
}

type gfySearchResult struct {
	Cursor  string                        `json:"cursor"`
	Gfycats []map[string]*json.RawMessage `json:"gfycats"`
}

func (p *gfycat) GetAttributionMessage() string {
	return "Powered by Gfycat"
}

// Return the URL of a GIF that matches the query, or an empty string if no GIF matches the query, or an error if the search failed
func (p *gfycat) GetGifURL(request string, cursor *string) (string, *model.AppError) {
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
	var pageCursor gfyPageCursor = gfyPageCursor{CursorForPage: "", PositionInPage: 0}
	if cursor != nil && *cursor != "" {
		if err := json.Unmarshal([]byte(*cursor), &pageCursor); err != nil {
			return "", p.errorGenerator.FromError("Could not unserialize Gfycat cursor", err)
		}
	}

	req, err := http.NewRequest("GET", baseURLGfycat+"/gfycats/search", nil)
	if err != nil {
		return "", p.errorGenerator.FromError("Could not generate GfyCat search URL", err)
	}
	q := req.URL.Query()
	q.Add("search_text", request)
	if pageCursor.CursorForPage != "" {
		q.Add("cursor", pageCursor.CursorForPage)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	r, err := p.httpClient.Do(req)
	if err != nil {
		return "", p.errorGenerator.FromError("Error calling the GfyCat search API", err)
	}

	if r.StatusCode != http.StatusOK {
		return "", p.errorGenerator.FromMessage(fmt.Sprintf("Error calling the GfyCat search API (HTTP Status: %v)", r.Status))
	}
	var response gfySearchResult
	decoder := json.NewDecoder(r.Body)
	if r.Body == nil {
		return "", p.errorGenerator.FromMessage("GfyCat search response body is empty")
	}
	if err = decoder.Decode(&response); err != nil {
		return "", p.errorGenerator.FromError("Could not parse Gfycat search response body", err)
	}
	if len(response.Gfycats) < 1 {
		return "", nil
	}
	gif := response.Gfycats[pageCursor.PositionInPage]
	urlNode, ok := gif[p.rendition]
	if !ok {
		return "", p.errorGenerator.FromMessage("No URL found for display style \"" + p.rendition + "\" in the response")
	}
	var url string
	if urlNode != nil {
		if err = json.Unmarshal(*urlNode, &url); err != nil {
			return "", p.errorGenerator.FromError("Could not read "+p.rendition+"node", err)
		}
	}
	// Ignore suffix without a Mattermost preview
	if url == "" || strings.HasSuffix(url, ".webm") || strings.HasSuffix(url, ".mp4") {
		urlNode, ok = gif["gifUrl"]
		if !ok {
			return "", p.errorGenerator.FromMessage("No URL found for the \"gifUrl\" in the response")
		}
		if err = json.Unmarshal(*urlNode, &url); err != nil {
			return "", p.errorGenerator.FromError("Could not read gifUrl node", err)
		}
	}
	if url == "" {
		return "", p.errorGenerator.FromMessage("An empty URL was returned for display style \"" + p.rendition + "\"")
	}

	noMoreResults := false
	if len(response.Gfycats) > pageCursor.PositionInPage+1 {
		// Next GIF will be in same page
		pageCursor.PositionInPage++
	} else {
		if response.Cursor != "" {
			// Next GIF will be in next page
			pageCursor.CursorForPage = response.Cursor
			pageCursor.PositionInPage = 0
		} else {
			noMoreResults = true
		}
	}
	if noMoreResults {
		*cursor = ""
	} else {
		nextCursor, err := json.Marshal(pageCursor)
		if err != nil {
			return "", p.errorGenerator.FromError("Could not serialize Gfycat cursor", err)
		}

		*cursor = string(nextCursor)
	}
	return url, nil
}
