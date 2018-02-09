package main

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/rpcplugin"
)

// Trigger used to define the slash command
const trigger = "gif"

// GiphyPlugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type GiphyPlugin struct {
	api           plugin.API
	configuration atomic.Value
	TeamId        string
	gifProvider   gifProvider
	enabled       bool
}

type GiphyPluginConfiguration struct {
	Rating    string
	Language  string
	Rendition string
}

// OnActivate register the plugin command
func (p *GiphyPlugin) OnActivate(api plugin.API) error {
	p.api = api
	p.enabled = true
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
	p.enabled = false
	return p.api.UnregisterCommand(p.TeamId, trigger)
}

// ExecuteCommand returns a post that displays a GIF choosen using Giphy
func (p *GiphyPlugin) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.enabled {
		return nil, appError("Cannot execute command while the plugin is disabled.", nil)
	}
	if p.api == nil {
		return nil, appError("Cannot access the plugin API.", nil)
	}
	cmd := "/" + trigger
	if strings.HasPrefix(args.Command, cmd) {
		keywords := strings.Replace(args.Command, cmd, "", 1)

		gifURL, err := p.gifProvider.getGifURL(p.config(), keywords)
		if err != nil {
			return nil, appError("Unable to get GIF URL", err)
		}

		text := " *[" + keywords + "](" + gifURL + ")*\n" + "![GIF for '" + keywords + "'](" + gifURL + ")"
		return &model.CommandResponse{ResponseType: "in_channel", Text: text}, nil
	}

	return nil, appError("Command trigger "+args.Command+"is not supported by this plugin.", nil)
}

func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError("Giphy Plugin", message, nil, errorMessage, http.StatusBadRequest)
}

// Install the RCP plugin
func main() {
	plugin := GiphyPlugin{}
	plugin.gifProvider = &giphyProvider{}
	rpcplugin.Main(&plugin)
}
