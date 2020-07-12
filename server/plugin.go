package main

import (
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/pkg/errors"
)

const (
	contextKeywords = "keywords"
	contextGifURL   = "gifURL"
	contextCursor   = "cursor"
	contextRootId   = "rootId"
)

// Plugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type Plugin struct {
	plugin.MattermostPlugin

	configurationLock sync.RWMutex
	configuration     *configuration

	gifProvider gifProvider
	httpHandler pluginHTTPHandler
	botId       string
}

// gifProvider exposes methods to get GIF URLs
type gifProvider interface {
	getGifURL(config *configuration, request string, cursor *string) (string, *model.AppError)
	getAttributionMessage() string
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(s string) (*http.Response, error)
}

var getGifProviderHttpClient = func() HttpClient {
	return http.DefaultClient
}

// OnActivate register the plugin commands
func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return errors.Wrap(err, "Could not load plugin configuration")
	}
	p.httpHandler = &defaultHTTPHandler{}
	return p.RegisterCommands()
}

// ExecuteCommand dispatch the command based on the trigger word
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if strings.HasPrefix(args.Command, "/"+triggerGifs) {
		return p.executeCommandGifShuffle(args.Command, args)
	}
	if strings.HasPrefix(args.Command, "/"+triggerGif) {
		return p.executeCommandGif(args.Command)
	}

	return nil, appError("Command trigger "+args.Command+"is not supported by this plugin.", nil)
}

// ServeHTTP serve the post actions for the shuffle command
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.handleHTTPRequest(w, r)
}

//appError generates a normalized error for this plugin
func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError(manifest.Name, message, nil, errorMessage, http.StatusBadRequest)
}
