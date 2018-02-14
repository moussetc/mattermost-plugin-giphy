package main

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/rpcplugin"
	"github.com/sanzaru/go-giphy"
)

const (
	// API key used to retrieve Giphy GIF
	giphyAPIKey = "dc6zaTOxFJmzC"
	// Trigger used to define the slash command
	trigger = "giphy"
)

// GiphyPlugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type GiphyPlugin struct {
	api           plugin.API
	configuration atomic.Value
	TeamId        string
}

type GiphyPluginConfiguration struct {
	Rating    string
	Language  string
	Rendition string
}

// OnActivate register the plugin command
func (p *GiphyPlugin) OnActivate(api plugin.API) error {
	p.api = api
	err := api.RegisterCommand(&model.Command{
		Trigger:          trigger,
		TeamId:           p.TeamId,
		Description:      "Displays a Giphy GIF that matches the keyword(s)",
		DisplayName:      "Giphy command",
		AutoComplete:     true,
		AutoCompleteDesc: "Displays a Giphy GIF that matches the keyword(s)",
		AutoCompleteHint: "happy kitty",
	})
	if err != nil {
		return err
	}
	return p.OnConfigurationChange()
}

func (p *GiphyPlugin) config() *GiphyPluginConfiguration {
	return p.configuration.Load().(*GiphyPluginConfiguration)
}

func (p *GiphyPlugin) OnConfigurationChange() error {
	var configuration GiphyPluginConfiguration
	err := p.api.LoadPluginConfiguration(&configuration)
	p.configuration.Store(&configuration)
	return err
}

// OnDeactivate unregisters the plugin command
func (p *GiphyPlugin) OnDeactivate() error {
	return p.api.UnregisterCommand(p.TeamId, trigger)
}

// ExecuteCommand returns a post that displays a GIF choosen using Giphy
func (p *GiphyPlugin) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cmd := "/" + trigger
	if strings.HasPrefix(args.Command, cmd) {
		keywords := strings.TrimLeft(args.Command, cmd+" ")

		gifURL, err := p.getGifURL(keywords)
		if err != nil {
			return nil, model.NewAppError("Giphy Plugin ExecuteCommand", "Could not get GIF", nil, "", http.StatusBadRequest)
		}

		text := "I searched for [" + keywords + "](" + gifURL + ") and found this:\n" + "![GIF for '" + keywords + "'](" + gifURL + ")"
		return &model.CommandResponse{ResponseType: "in_channel", Text: text, Username: args.UserId}, nil
	}

	return nil, model.NewAppError("Giphy Plugin ExecuteCommand", "Expected trigger "+cmd+" but got "+args.Command, nil, "", http.StatusBadRequest)
}

// getGifURL return the URL of a small Giphy GIF that more or less correspond to requested keywords
func (p *GiphyPlugin) getGifURL(request string) (string, error) {
	giphy := libgiphy.NewGiphy(giphyAPIKey)
	config := p.config()

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

// Install the RCP plugin
func main() {
	rpcplugin.Main(&GiphyPlugin{})
}
