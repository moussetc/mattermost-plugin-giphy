package main

import (
	"github.com/sanzaru/go-giphy"
)

// API key used to retrieve Giphy GIF
const giphyAPIKey = "dc6zaTOxFJmzC"

// gifProvider exposes methods to get GIF URLs
type gifProvider interface {
	getGifURL(config *GiphyPluginConfiguration, request string) (string, error)
}

// giphyProvider get GIF URLs from the Giphy API
type giphyProvider struct{}

// getGifURL return the URL of a small Giphy GIF that more or less correspond to requested keywords
func (*giphyProvider) getGifURL(config *GiphyPluginConfiguration, request string) (string, error) {
	giphy := libgiphy.NewGiphy(giphyAPIKey)

	data, err := giphy.GetTranslate(request, config.Rating, config.Language, false)
	if err != nil {
		return "", err
	}
	switch config.Rendition {
	case "fixed_height":
		return data.Data.Images.Fixed_height.Url, nil
	case "fixed_height_still":
		return data.Data.Images.Fixed_height_still.Url, nil
	case "fixed_height_small":
		return data.Data.Images.Fixed_height_small.Url, nil
	case "fixed_height_small_still":
		return data.Data.Images.Fixed_height_small_still.Url, nil
	case "fixed_width":
		return data.Data.Images.Fixed_width.Url, nil
	case "fixed_width_still":
		return data.Data.Images.Fixed_width_still.Url, nil
	case "fixed_width_small":
		return data.Data.Images.Fixed_width_small.Url, nil
	case "fixed_width_small_still":
		return data.Data.Images.Fixed_width_small_still.Url, nil
	case "downsized":
		return data.Data.Images.Downsized.Url, nil
	case "downsized_large":
		return data.Data.Images.Downsized_large.Url, nil
	case "downsized_still":
		return data.Data.Images.Downsized_still.Url, nil
	case "original":
		return data.Data.Images.Original.Url, nil
	case "original_still":
		return data.Data.Images.Original_still.Url, nil
	}
	return data.Data.Images.Fixed_height_small.Url, nil
}
