package main

import (
	"fmt"
	"encoding/json"
	"bytes"
	"net/http"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/model"
)

// gifyCatProvider get GIF URLs from the GfyCat API, using Mattermost settings
type gfyCatProvider struct{}

const (
	GFYCAT_BASE_URL = "https://api.gfycat.com/v1"
)

type authResponse struct {
	AccessToken	string  `json:"access_token"`
}

func getAuthentToken(api *plugin.API) (string, *model.AppError) {
	clientId := *(*api).GetConfig().ServiceSettings.GfycatApiKey
	clientSecret := *(*api)	.GetConfig().ServiceSettings.GfycatApiSecret
	if clientId == "" || clientSecret == "" {
		return "", appError("GfyCat API key and secrets must be defined in the System Console", nil)
	}
	values := map[string]string{"grant_type": "client_credentials", "client_id": clientId, "client_secret": clientSecret}
	jsonValue, _ := json.Marshal(values)

	req, err := http.NewRequest("POST", GFYCAT_BASE_URL+"/oauth/token", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", appError("Could not generate authentication URL", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", appError("Could not get GfyCat auth token", err)
	}
	if r.StatusCode != http.StatusOK {
		return "", appError("Could not get GfyCat auth token (HTTP Status: "+ fmt.Sprintf("%v", r.StatusCode)+")", nil)
	}

	var response authResponse
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return "", appError("Could not parse Gfycat authent response body", err)
	}
	if response.AccessToken == "" {
		return "", appError("An empty GfyCat auth token was returned", err)
	}
	return response.AccessToken, nil
}

type gfySearchResult struct {
	Cursor string	`json:"cursor"`
	Gfycats	[]gfyGIF	`json:"gfycats"`
}

type gfyGIF struct {
	GifUrl string	`json:"gifUrl"`
}

// getGifURL return the URL of a GIF that matches the requested keywords
func (p *gfyCatProvider) getGifURL(api *plugin.API, config *GiphyPluginConfiguration, request string, counter int) (string, error) {
	// TODO : only get a new token if the previous one is expired (where to stock it ? beware thread)
	token, authErr := getAuthentToken(api)
	if authErr != nil {
		return "", authErr
	}

	req, err := http.NewRequest("GET", GFYCAT_BASE_URL+"/gfycats/search", nil)
	if err != nil {
		return "", appError("Could not generate GfyCat search URL", err)
	}
// TODO there's a "cursor" which might serve the same than the Giphy counter, so it's not always the same GIF returned!
	q := req.URL.Query()
	q.Add("search_text", request)
	q.Add("count", "1")
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization", token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", appError("Error calling the GfyCat search API", err)
	}

	if r.StatusCode != http.StatusOK {
		return "", appError("Error calling the GfyCat search API (HTTP Status: "+string(r.StatusCode)+")", nil)
	}
	var response gfySearchResult
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&response); err != nil {
		return "", appError("Could not parse Gfycat search response body", err)
	}
	if len(response.Gfycats) < 1 {
		return "", appError("An empty list of GIFs was returned", err)
	}
	return response.Gfycats[0].GifUrl, nil
}
