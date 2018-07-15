package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/rpcplugin"
)

const (
	// Triggers used to define slash commands
	triggerGif         = "gif"
	triggerGifs        = "gifs"
	pluginPath  string = "/plugins/com.github.moussetc.mattermost.plugin.giphy"
	actionURL   string = "/action"
	rootURL     string = "http://localhost:8065"
)

// GiphyPlugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type GiphyPlugin struct {
	api           plugin.API
	configuration atomic.Value
	gifProvider   gifProvider
	router        *mux.Router
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
		Description:      "Shows a preview of several GIFS matching the keyword(s)",
		DisplayName:      "Giphy preview command",
		AutoComplete:     true,
		AutoCompleteDesc: "Shows a preview of several GIFS matching the keyword(s)",
		AutoCompleteHint: "happy kitty",
	})
	if err != nil {
		return err
	}

	// Serve URL for TODO???
	p.router = mux.NewRouter()
	p.router.HandleFunc(actionURL, p.handleAction)

	return p.OnConfigurationChange()
}

func (p *GiphyPlugin) handleAction(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	channelID := r.URL.Query().Get("channelId")
	gifURL := r.URL.Query().Get("gifUrl")
	keywords := r.URL.Query().Get("keyword")

	post := &model.Post{
		Message:   " *[" + keywords + "](" + gifURL + ")*\n" + "![GIF for '" + keywords + "'](" + gifURL + ")",
		ChannelId: channelID,
		UserId:    userID,
	}

	if _, err := p.api.CreatePost(post); err != nil {
		fmt.Fprint(w, "Error: "+err.Message)
		return
	}
	fmt.Fprint(w, "The GIF was posted publicly, you can close this tab now (Ctrl+W). Have a good day!")
}

func (p *GiphyPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Mattermost-User-Id") == "" {
		http.Error(w, "please log in", http.StatusForbidden)
		return
	}

	p.router.ServeHTTP(w, r)
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

// OnDeactivate handles plugin deactivation
func (p *GiphyPlugin) OnDeactivate() error {
	p.enabled = false
	return nil
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
		return p.executeCommandGifs(args)
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
func (p *GiphyPlugin) executeCommandGifs(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	keywords := getCommandKeywords(args.Command, triggerGifs)
	gifURLs, err := p.gifProvider.getMultipleGifsURL(p.config(), keywords)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	tableHeader := "|"
	tableSeparator := "|"
	tableRow := "|"
	for _, gifURL := range gifURLs {
		tableSeparator += ":----:|"
		tableRow += fmt.Sprintf("[![GIF for '%s'](%s)](%s)", keywords, gifURL, gifURL) + "|"

		actionURL, err := url.Parse(rootURL + pluginPath + actionURL)
		if err != nil {
			return nil, appError("Unable to build action URL, make sure the server root URL is configured", err)
		}

		params := url.Values{}
		params.Add("channelId", args.ChannelId)
		params.Add("userId", args.UserId)
		params.Add("keyword", keywords)
		params.Add("gifUrl", gifURL)
		actionURL.RawQuery = params.Encode()
		tableHeader += "[Chose me](" + actionURL.String() + ")|"
	}
	text := fmt.Sprintf("%s\n%s\n%s\n *Suggestions for '%s'*", tableHeader, tableSeparator, tableRow, keywords)
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
