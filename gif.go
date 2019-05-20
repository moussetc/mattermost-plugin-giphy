package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// gifProvider exposes methods to get GIF URLs
type gifProvider interface {
	getGifURL(config *GiphyPluginConfiguration, request string) (string, error)
}

// giphyProvider get GIF URLs from the Giphy API without any external, out-of-date library
type giphyProvider struct{}

const (
	BASE_URL = "http://api.giphy.com/v1/gifs"
)

// getGifURL return the URL of a GIF that matches the requested keywords
func (p *giphyProvider) getGifURL(config *GiphyPluginConfiguration, request string) (string, error) {
	if config.APIKey == "" {
		return "", appError("Giphy API key is empty", nil)
	}
	req, err := http.NewRequest("GET", BASE_URL+"/translate", nil)
	if err != nil {
		return "", appError("Could not generate URL", err)
	}

	q := req.URL.Query()
	q.Add("api_key", config.APIKey)
	q.Add("s", request)
	q.Add("weirdness", "0")

	req.URL.RawQuery = q.Encode()
	requestURL := req.URL.String()

	r, err := http.DefaultClient.Get(requestURL)
	if err != nil {
		return "", appError("Error calling the Giphy API", err)
	}

	if r.StatusCode != http.StatusOK {
		return "", appError("Error calling the Giphy API (HTTP Status: "+string(r.StatusCode)+")", nil)
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", appError("Error reading the Giphy API response", err)
	}

	var rootNode map[string]*json.RawMessage
	err = json.Unmarshal(data, &rootNode)
	if err != nil {
		return "", appError("Error unmarshalling JSON", err)
	}
	var dataNode map[string]*json.RawMessage
	err = json.Unmarshal(*rootNode["data"], &dataNode)
	if err != nil {
		return "", appError("Error unmarshalling JSON", err)
	}

	var imagesNode map[string]*json.RawMessage
	err = json.Unmarshal(*dataNode["images"], &imagesNode)
	if err != nil {
		return "", appError("Error unmarshalling JSON", err)
	}

	var imageNode map[string]*json.RawMessage
	err = json.Unmarshal(*imagesNode[config.Rendition], &imageNode)
	if err != nil {
		return "", appError("Error unmarshalling JSON", err)
	}

	var urlNode string
	err = json.Unmarshal(*imageNode["url"], &urlNode)
	if err != nil {
		return "", appError("Error unmarshalling JSON", err)
	}
	return urlNode, nil
}
