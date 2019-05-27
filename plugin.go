package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

const (
	// Triggers used to define slash commands
	triggerGif      = "gif"
	triggerGifs     = "gifs"
	pluginID        = "com.github.moussetc.mattermost.plugin.giphy" // TODO get that from manifest
	contextKeywords = "keywords"
	contextGifURL   = "gifURL"
	contextCursor   = "cursor"
	URLShuffle      = "/shuffle"
	URLCancel       = "/cancel"
	URLSend         = "/send"
)

// Plugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type Plugin struct {
	plugin.MattermostPlugin
	siteURL string

	configuration atomic.Value
	gifProvider   gifProvider
	enabled       bool
}

// PluginConfiguration contains all plugin parameters
type PluginConfiguration struct {
	Provider        string
	Rating          string
	Language        string
	Rendition       string
	RenditionGfycat string
	APIKey          string
}

// gifProvider exposes methods to get GIF URLs
type gifProvider interface {
	getGifURL(API *plugin.API, config *PluginConfiguration, request string, cursor *string) (string, error)
}

// OnActivate register the plugin commands
func (p *Plugin) OnActivate() error {
	if p.API.GetConfig().ServiceSettings.SiteURL == nil {
		return appError("siteURL must be set for the plugin to work. Please set a siteURL and restart the plugin", nil)
	}
	p.siteURL = *p.API.GetConfig().ServiceSettings.SiteURL

	if err := p.OnConfigurationChange(); err != nil {
		return appError("Could not load plugin configuration", err)
	}
	p.enabled = true
	err := p.API.RegisterCommand(&model.Command{
		Trigger:          triggerGif,
		Description:      "Post a GIF matching your search",
		DisplayName:      "Giphy Search",
		AutoComplete:     true,
		AutoCompleteDesc: "Post a GIF matching your search",
		AutoCompleteHint: "happy kitty",
	})
	if err != nil {
		return err
	}
	err = p.API.RegisterCommand(&model.Command{
		Trigger:          triggerGifs,
		Description:      "Preview a GIF",
		DisplayName:      "Giphy Shuffle",
		AutoComplete:     true,
		AutoCompleteDesc: "Let you preview and shuffle a GIF before posting for real",
		AutoCompleteHint: "mayhem guy",
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Plugin) config() *PluginConfiguration {
	return p.configuration.Load().(*PluginConfiguration)
}

// OnConfigurationChange apply a new plugin configuration
func (p *Plugin) OnConfigurationChange() error {
	var configuration PluginConfiguration
	if err := p.API.LoadPluginConfiguration(&configuration); err != nil {
		return err
	}
	if configuration.Provider == "" {
		return appError("GIF Provider setting must be set", nil)
	}
	if configuration.Provider == "giphy" {
		if configuration.APIKey == "" {
			return appError("The API Key setting must be set for Giphy", nil)
		}
		p.gifProvider = &giphyProvider{}
	} else {
		p.gifProvider = &gfyCatProvider{}
	}
	p.configuration.Store(&configuration)
	return nil
}

// OnDeactivate handles plugin deactivation
func (p *Plugin) OnDeactivate() error {
	p.enabled = false
	return nil
}

// ExecuteCommand dispatch the command based on the trigger word
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.enabled {
		return nil, appError("Cannot execute command while the plugin is disabled.", nil)
	}
	if p.API == nil {
		return nil, appError("Cannot access the plugin API.", nil)
	}
	if strings.HasPrefix(args.Command, "/"+triggerGifs) {
		return p.executeCommandGifShuffle(args.Command, args)
	}
	if strings.HasPrefix(args.Command, "/"+triggerGif) {
		return p.executeCommandGif(args.Command)
	}

	return nil, appError("Command trigger "+args.Command+"is not supported by this plugin.", nil)
}

func getCommandKeywords(commandLine string, trigger string) string {
	return strings.Replace(commandLine, "/"+trigger, "", 1)
}

// ServeHTTP serve the post action to display an ephemeral spoiler
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case URLShuffle:
		p.handleHTTPAction(p.handleShuffle, c, w, r)
	case URLSend:
		p.handleHTTPAction(p.handlePost, c, w, r)
	case URLCancel:
		p.handleHTTPAction(p.handleCancel, c, w, r)
	default:
		http.NotFound(w, r)
	}
}

// Generate an attachment for an action Button that will point to a plugin HTTP handler
func (p *Plugin) generateButton(name string, urlAction string, context map[string]interface{}) *model.PostAction {
	return &model.PostAction{
		Name: name,
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL:     fmt.Sprintf("%s/plugins/%s"+urlAction, p.siteURL, pluginID),
			Context: context,
		},
	}
}

//appError generates a normalized error for this plugin
func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError("GIF Plugin", message, nil, errorMessage, http.StatusBadRequest)
}

// Install the RCP plugin
func main() {
	p := Plugin{}
	plugin.ClientMain(&p)
}
