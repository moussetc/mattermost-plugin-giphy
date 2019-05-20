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
	URLShuffle      = "/shuffle"
	URLCancel       = "/cancel"
	URLSend         = "/send"
)

// GiphyPlugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type GiphyPlugin struct {
	plugin.MattermostPlugin
	siteURL string

	configuration atomic.Value
	gifProvider   gifProvider
	enabled       bool
}

// GiphyPluginConfiguration contains all plugin parameters
type GiphyPluginConfiguration struct {
	Rating        string
	Language      string
	Rendition     string
	APIKey        string
	EncryptionKey string
}

// OnActivate register the plugin commands
func (p *GiphyPlugin) OnActivate() error {
	if p.API.GetConfig().ServiceSettings.SiteURL == nil {
		return appError("siteURL must be set for the plugin to work. Please set a siteURL and restart the plugin", nil)
	}
	p.siteURL = *p.API.GetConfig().ServiceSettings.SiteURL

	p.enabled = true
	err := p.API.RegisterCommand(&model.Command{
		Trigger:          triggerGif,
		Description:      "Post a GIF from Giphy matching your search",
		DisplayName:      "Giphy Search",
		AutoComplete:     true,
		AutoCompleteDesc: "Post a GIF from Giphy matching your search",
		AutoCompleteHint: "happy kitty",
	})
	if err != nil {
		return err
	}
	err = p.API.RegisterCommand(&model.Command{
		Trigger:          triggerGifs,
		Description:      "Preview a GIF from Giphy",
		DisplayName:      "Giphy Shuffle",
		AutoComplete:     true,
		AutoCompleteDesc: "Let you preview and shuffle a GIF from Giphy before posting for real",
		AutoCompleteHint: "mayhem guy",
	})
	if err != nil {
		return err
	}
	return p.OnConfigurationChange()
}

func (p *GiphyPlugin) config() *GiphyPluginConfiguration {
	return p.configuration.Load().(*GiphyPluginConfiguration)
}

// OnConfigurationChange apply a new plugin configuration
func (p *GiphyPlugin) OnConfigurationChange() error {
	var configuration GiphyPluginConfiguration
	err := p.API.LoadPluginConfiguration(&configuration)
	p.configuration.Store(&configuration)
	return err
}

// OnDeactivate handles plugin deactivation
func (p *GiphyPlugin) OnDeactivate() error {
	p.enabled = false
	return nil
}

// ExecuteCommand dispatch the command based on the trigger word
func (p *GiphyPlugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
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
func (p *GiphyPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
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
func (p *GiphyPlugin) generateButton(name string, urlAction string, context map[string]interface{}) *model.PostAction {
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
	return model.NewAppError("Giphy Plugin", message, nil, errorMessage, http.StatusBadRequest)
}

// Install the RCP plugin
func main() {
	p := GiphyPlugin{}
	p.gifProvider = &giphyProvider{}
	plugin.ClientMain(&p)
}
