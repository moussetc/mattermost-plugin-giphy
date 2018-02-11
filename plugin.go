package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/rpcplugin"
)

const (
	// Triggers used to define slash commands
	triggerGif  = "gif"
	triggerGifs = "gifs"
)

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
	APIKey    string
}

// OnActivate register the plugin commands
func (p *GiphyPlugin) OnActivate(api plugin.API) error {
	p.api = api
	p.enabled = true
	err := api.RegisterCommand(&model.Command{
		Trigger:          triggerGif,
		TeamId:           p.TeamId,
		Description:      "Posts a Giphy GIF that matches the keyword(s)",
		DisplayName:      "Giphy command",
		AutoComplete:     true,
		AutoCompleteDesc: "Posts a Giphy GIF that matches the keyword(s)",
		AutoCompleteHint: "happy kitty",
	})
	if err != nil {
		return err
	}

	err = api.RegisterCommand(&model.Command{
		Trigger:          triggerGifs,
		TeamId:           p.TeamId,
		Description:      "Shows a preview of 10 GIFS matching the keyword(s)",
		DisplayName:      "Giphy preview command",
		AutoComplete:     true,
		AutoCompleteDesc: "Shows a preview of 10 GIFS matching the keyword(s)",
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

// OnDeactivate unregisters the plugin commands
func (p *GiphyPlugin) OnDeactivate() error {
	p.enabled = false
	err := p.api.UnregisterCommand(p.TeamId, triggerGif)
	if err != nil {
		return err
	}
	return p.api.UnregisterCommand(p.TeamId, triggerGifs)
}

// ExecuteCommand returns a post that displays a GIF choosen using Giphy
func (p *GiphyPlugin) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.enabled {
		return nil, appError("Cannot execute command while the plugin is disabled.", nil)
	}
	if p.api == nil {
		return nil, appError("Cannot access the plugin API.", nil)
	}
	if strings.HasPrefix(args.Command, "/"+triggerGifs) {
		return p.executeCommandGifs(args.Command)
	}
	if strings.HasPrefix(args.Command, "/"+triggerGif) {
		return p.executeCommandGif(args.Command)
	}

	return nil, appError("Command trigger "+args.Command+"is not supported by this plugin.", nil)
}

// executeCommandGif returns a public post containing a matching GIF
func (p *GiphyPlugin) executeCommandGif(command string) (*model.CommandResponse, *model.AppError) {
	keywords := getCommandKeywords(command, triggerGif)
	gifURL, err := p.gifProvider.getGifURL(p.config(), keywords)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	text := " *[" + keywords + "](" + gifURL + ")*\n" + "![GIF for '" + keywords + "'](" + gifURL + ")"
	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, Text: text}, nil
}

// executeCommandGif returns a private post containing a list of matching GIFs
func (p *GiphyPlugin) executeCommandGifs(command string) (*model.CommandResponse, *model.AppError) {
	keywords := getCommandKeywords(command, triggerGifs)
	gifURLs, err := p.gifProvider.getMultipleGifsURL(p.config(), keywords)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	text := fmt.Sprintf(" *Suggestions for '%s':*", keywords)
	for i, url := range gifURLs {
		if i > 0 {
			text += "\t"
		}
		text += fmt.Sprintf("[![GIF for '%s'](%s)](%s)", keywords, url, url)
	}
	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL, Text: text}, nil
}

func getCommandKeywords(commandLine string, trigger string) string {
	return strings.Replace(commandLine, "/"+trigger, "", 1)
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
