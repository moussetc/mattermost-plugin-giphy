package main

import (
	"errors"

	"github.com/sanzaru/go-giphy"
)

// gifProvider exposes methods to get GIF URLs
type gifProvider interface {
	getGifURL(config *GiphyPluginConfiguration, request string) (string, error)
	getMultipleGifsURL(config *GiphyPluginConfiguration, request string) ([]string, error)
}

// giphyProvider get GIF URLs from the Giphy API
type giphyProvider struct{}

// getGifURL return the URL of a small Giphy GIF that more or less correspond to requested keywords
func (p *giphyProvider) getGifURL(config *GiphyPluginConfiguration, request string) (string, error) {
	if config.APIKey == "" {
		return "", errors.New("Giphy API key is empty")
	}

	giphy := libgiphy.NewGiphy(config.APIKey)

	response, err := giphy.GetTranslate(request, config.Rating, config.Language, false)
	if err != nil {
		return "", err
	}

	data := giphyData{
		FixedHeight:            gif{URL: response.Data.Images.Fixed_height.Url},
		FixedHeightStill:       gif{URL: response.Data.Images.Fixed_height_still.Url},
		FixedHeightDownsampled: gif{URL: response.Data.Images.Fixed_height_downsampled.Url},
		FixedWidth:             gif{URL: response.Data.Images.Fixed_width.Url},
		FixedWidthStill:        gif{URL: response.Data.Images.Fixed_width_still.Url},
		FixedWidthDownsampled:  gif{URL: response.Data.Images.Fixed_width_downsampled.Url},
		FixedHeightSmall:       gif{URL: response.Data.Images.Fixed_height_small.Url},
		FixedHeightSmallStill:  gif{URL: response.Data.Images.Fixed_height_small_still.Url},
		FixedWidthSmall:        gif{URL: response.Data.Images.Fixed_width_small.Url},
		FixedWidthSmallStill:   gif{URL: response.Data.Images.Fixed_width_small_still.Url},
		Downsized:              gif{URL: response.Data.Images.Downsized.Url},
		DownsizedStill:         gif{URL: response.Data.Images.Downsized_still.Url},
		DownsizedLarge:         gif{URL: response.Data.Images.Downsized_large.Url},
		Original:               gif{URL: response.Data.Images.Original.Url},
		OriginalStill:          gif{URL: response.Data.Images.Original_still.Url},
	}
	return p.getGifForRendition(config.Rendition, &data).URL, nil
}

// getGifURL return the URL of a small Giphy GIF that more or less correspond to requested keywords
func (p *giphyProvider) getMultipleGifsURL(config *GiphyPluginConfiguration, request string) ([]string, error) {
	if config.APIKey == "" {
		return nil, errors.New("Giphy API key is empty")
	}

	giphy := libgiphy.NewGiphy(config.APIKey)
	response, err := giphy.GetSearch(request, 5, 0, config.Rating, config.Language, false)
	if err != nil {
		return nil, err
	}
	urls := make([]string, len(response.Data))
	for _, data := range response.Data {
		data := giphyData{
			FixedHeight:            gif{URL: data.Images.Fixed_height.Url},
			FixedHeightStill:       gif{URL: data.Images.Fixed_height_still.Url},
			FixedHeightDownsampled: gif{URL: data.Images.Fixed_height_downsampled.Url},
			FixedWidth:             gif{URL: data.Images.Fixed_width.Url},
			FixedWidthStill:        gif{URL: data.Images.Fixed_width_still.Url},
			FixedWidthDownsampled:  gif{URL: data.Images.Fixed_width_downsampled.Url},
			FixedHeightSmall:       gif{URL: data.Images.Fixed_height_small.Url},
			FixedHeightSmallStill:  gif{URL: data.Images.Fixed_height_small_still.Url},
			FixedWidthSmall:        gif{URL: data.Images.Fixed_width_small.Url},
			FixedWidthSmallStill:   gif{URL: data.Images.Fixed_width_small_still.Url},
			Downsized:              gif{URL: data.Images.Downsized.Url},
			DownsizedStill:         gif{URL: data.Images.Downsized_still.Url},
			DownsizedLarge:         gif{URL: data.Images.Downsized_large.Url},
			Original:               gif{URL: data.Images.Original.Url},
			OriginalStill:          gif{URL: data.Images.Original_still.Url},
		}
		urls = append(urls, p.getGifForRendition(config.Rendition, &data).URL)
	}
	return urls, nil
}

type gif struct {
	URL string
}

type giphyData struct {
	FixedHeight            gif
	FixedHeightStill       gif
	FixedHeightDownsampled gif
	FixedWidth             gif
	FixedWidthStill        gif
	FixedWidthDownsampled  gif
	FixedHeightSmall       gif
	FixedHeightSmallStill  gif
	FixedWidthSmall        gif
	FixedWidthSmallStill   gif
	Downsized              gif
	DownsizedStill         gif
	DownsizedLarge         gif
	Original               gif
	OriginalStill          gif
}

// getGifURL return the URL of a small Giphy GIF that more or less correspond to requested keywords
func (*giphyProvider) getGifForRendition(renditionStyle string, data *giphyData) gif {
	var gif gif
	switch renditionStyle {
	case "fixed_height":
		gif = data.FixedHeight
	case "fixed_height_still":
		gif = data.FixedHeightStill
	case "fixed_height_small":
		gif = data.FixedHeightSmall
	case "fixed_height_small_still":
		gif = data.FixedHeightSmallStill
	case "fixed_width":
		gif = data.FixedWidth
	case "fixed_width_still":
		gif = data.FixedWidthStill
	case "fixed_width_small":
		gif = data.FixedWidthSmall
	case "fixed_width_small_still":
		gif = data.FixedWidthSmallStill
	case "downsized":
		gif = data.Downsized
	case "downsized_large":
		gif = data.DownsizedLarge
	case "downsized_still":
		gif = data.DownsizedStill
	case "original":
		gif = data.Original
	case "original_still":
		gif = data.OriginalStill
	default:
		gif = data.FixedHeightSmall
	}
	return gif
}
