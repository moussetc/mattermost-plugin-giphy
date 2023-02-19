package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v6/model"
)

// giphy find GIFs using the giphy API
type giphy struct {
	abstractGifProvider
	apiKey  string
	rootURL string
}

const (
	baseURLGiphy = "https://api.giphy.com/v1/gifs"
)

type GiphyData struct {
	Images map[string]struct {
		URL string `json:"url"`
	} `json:"images"`
}

type GiphySearchResult struct {
	Data       []GiphyData `json:"data"`
	Pagination struct {
		Offset int `json:"offset"`
	} `json:"pagination"`
}

type GiphyRandomResult struct {
	Data GiphyData `json:"data"`
}

type GiphyRandomEmptyResult struct {
	Data []GiphyData `json:"data"`
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
func (p *giphy) GetGifURL(request string, cursor *string, random bool) ([]string, *model.AppError) {
	if random {
		return p.getRandomGifURL(request)
	}
	return p.getSearchGifURL(request, cursor)
}

// Return the URL of a GIF that matches the query, or an empty string if no GIF matches the query, or an error if the search failed
func (p *giphy) getSearchGifURL(request string, cursor *string) ([]string, *model.AppError) {
	parameters := map[string]string{"q": request}
	if counter, err2 := strconv.Atoi(*cursor); err2 == nil {
		parameters["offset"] = fmt.Sprintf("%d", counter)
	}
	if len(p.language) > 0 {
		parameters["lang"] = p.language
	}

	body, err := p.callGiphyEndpoint("search", parameters)
	if err != nil {
		return []string{}, err
	}

	var response GiphySearchResult
	if decodeErr := json.Unmarshal(body, &response); decodeErr != nil {
		return []string{}, p.errorGenerator.FromError("Could not parse Giphy response body", decodeErr)
	}

	if len(response.Data) < 1 {
		return []string{}, nil
	}

	urls := []string{}
	for i := range response.Data {
		if url, err := p.getURL(response.Data[i]); err == nil {
			urls = append(urls, url)
		}
	}

	if len(urls) < 1 {
		return []string{}, p.errorGenerator.FromMessage("No gifs found for display style \"" + p.rendition + "\" in the response")
	}

	*cursor = fmt.Sprintf("%d", response.Pagination.Offset+1)

	return urls, nil
}

// Return the URL of a random GIF that matches the query, or an empty string if no GIF matches the query, or an error if the search failed
func (p *giphy) getRandomGifURL(request string) ([]string, *model.AppError) {
	body, err := p.callGiphyEndpoint("random", map[string]string{"tag": request})
	if err != nil {
		return []string{}, err
	}

	var response GiphyRandomResult
	if err := json.Unmarshal(body, &response); err != nil {
		// Giphy API has the bad taste to return a different structure in case of no GIF found...
		var emptyResponse GiphyRandomEmptyResult
		if err = json.Unmarshal(body, &emptyResponse); err == nil {
			// No GIF found
			return []string{}, nil
		}
		return []string{}, p.errorGenerator.FromError("Could not parse Giphy response body", err)
	}

	url, err := p.getURL(response.Data)
	if err != nil {
		return []string{}, err
	}
	return []string{url}, nil
}

func (p *giphy) callGiphyEndpoint(endpoint string, customParameters map[string]string) ([]byte, *model.AppError) {
	req, err := http.NewRequest("GET", baseURLGiphy+"/"+endpoint, nil)
	if err != nil {
		return nil, p.errorGenerator.FromError("Could not generate URL", err)
	}

	q := req.URL.Query()

	q.Add("api_key", p.apiKey)
	if p.rating != "none" && len(p.rating) > 0 {
		q.Add("rating", p.rating)
	}
	for key, value := range customParameters {
		q.Add(key, value)
	}

	req.URL.RawQuery = q.Encode()

	r, err := p.httpClient.Do(req)
	if err != nil {
		return nil, p.errorGenerator.FromError("Error calling the Giphy API "+req.URL.RawQuery, err)
	}
	if r.Body != nil {
		defer r.Body.Close()
	}

	if r.StatusCode != http.StatusOK {
		explanation := ""
		if r.StatusCode == http.StatusTooManyRequests {
			explanation = ", this can happen if you're using the default Giphy API key"
		}
		return nil, p.errorGenerator.FromMessage(fmt.Sprintf("Error calling the Giphy API (HTTP Status: %v%s)", r.Status, explanation))
	}
	if r.Body == nil {
		return nil, p.errorGenerator.FromMessage("Giphy response body is empty")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, p.errorGenerator.FromError("Unable to read response body", err)
	}
	return body, nil
}

func (p *giphy) getURL(gif GiphyData) (string, *model.AppError) {
	url := gif.Images[p.rendition].URL

	if len(url) < 1 {
		return "", p.errorGenerator.FromMessage("No URL found for display style \"" + p.rendition + "\" in the response")
	}
	return url, nil
}
