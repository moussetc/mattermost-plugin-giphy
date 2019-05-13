package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/model"
	//"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/plugin"
)

const (
	// Triggers used to define slash commands
	triggerGif  = "gif"
	triggerGifs = "gifs"
	pluginID = "com.github.moussetc.mattermost.plugin.giphy" // TODO get that from manifest
)

// GiphyPlugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type GiphyPlugin struct {
	plugin.MattermostPlugin
	siteURL       string

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
		Description:      "Posts a Giphy GIF that matches the keyword(s)",
		DisplayName:      "Giphy command",
		AutoComplete:     true,
		AutoCompleteDesc: "Posts a Giphy GIF that matches the keyword(s)",
		AutoCompleteHint: "happy kitty",
	})
	if err != nil {
		return err
	}
	err = p.API.RegisterCommand(&model.Command{
		Trigger:          triggerGifs,
		Description:      "TODO shuffle",// TODO update that and also update the README!!
		DisplayName:      "Giphy command, shuffle mode",
		AutoComplete:     true,
		AutoCompleteDesc: "Posts a Giphy GIF that matches the keyword(s)",
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

// ExecuteCommand returns a post that displays a GIF choosen using Giphy
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

// executeCommandGifShuffle returns an ephemeral (private) post with one GIF that can either be posted, shuffled or canceled
func (p *GiphyPlugin) executeCommandGifShuffle(command string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	keywords := getCommandKeywords(command, triggerGifs)
	gifURL, err := p.gifProvider.getGifURL(p.config(), keywords)
	if err != nil {
		return nil, appError("Unable to get GIF URL", err)
	}

	actions := []*model.PostAction{}
	actions = append(actions, &model.PostAction{
		Name: "Cancel",
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/%s/cancel", p.siteURL, pluginID),
		},
	})
	actions = append(actions, &model.PostAction{
		Name: "Shuffle",
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/%s/shuffle", p.siteURL, pluginID),
		},
	})
	actions = append(actions, &model.PostAction{
		Name: "Post",
		Type: model.POST_ACTION_TYPE_BUTTON,
		Integration: &model.PostActionIntegration{
			URL: fmt.Sprintf("%s/plugins/%s/post", p.siteURL, pluginID),
		},
	})


	attachments := []*model.SlackAttachment{}
	attachments = append(attachments, &model.SlackAttachment{
		Text : " *[" + keywords + "](" + gifURL + ")*\n" + "![GIF for '" + keywords + "'](" + gifURL + ")",
		Actions: actions,
	})

	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL, Attachments: attachments}, nil
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

// ServeHTTP serve the post action to display an ephemeral spoiler
func (p *GiphyPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogWarn("URL = "+r.URL.Path)
	p.API.LogWarn("Context = "+ fmt.Sprintf("%+v\n", c))
	p.API.LogWarn("Request = "+ fmt.Sprintf("%+v\n", r))
	switch r.URL.Path {
	case "/shuffle":
		p.handleShuffle(w, r)
	case "/post":
		p.handlePost(w, r)
	case "/cancel":
		p.handleCancel(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleCancel delete the ephemeral shuffle post
func (p *GiphyPlugin) handleCancel(w http.ResponseWriter, r *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	post := &model.Post{
		Id: request.PostId,
	}
	p.API.DeleteEphemeralPost(request.UserId, post)

	/*response := &model.PostActionIntegrationResponse{}
	response.Update = post
	w.Header().Set("Content-Type", "application/json")*/
	w.WriteHeader(http.StatusOK)
	//_, _ = w.Write(response.ToJson())
}

// handleShuffle replace the GIF in the ephemeral shuffle post by a new one
func (p *GiphyPlugin) handleShuffle(w http.ResponseWriter, r *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO put the keywords into action context, then do another giphy search and set message accordingly
	post := &model.Post{
		Id: request.PostId,
		Message: "YOLOOOOO",
	}
	p.API.UpdateEphemeralPost(request.UserId, post)

	w.WriteHeader(http.StatusOK)
}

// handlePost post the actual GIF and delete the obsolete ephemeral post
func (p *GiphyPlugin) handlePost(w http.ResponseWriter, r *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(r.Body)
	if request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ephemeralPost := &model.Post{
		Id: request.PostId,
	}
	p.API.DeleteEphemeralPost(request.UserId, ephemeralPost)
	// TODO get context data and recreate content from that
	post := &model.Post{
		Message: "TODOOOOOOOOOO",
		UserId: request.UserId,
		ChannelId: request.ChannelId,
	}
	_, err := p.API.CreatePost(post)
	if err != nil {
		p.API.LogError("Could not create post", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// TODO : prevent those freaking error messages in the logs "Action integration error err=EOF"
	// even if they don't actually prevent the plugin from working...
}

// Install the RCP plugin
func main() {
	p := GiphyPlugin{}
	p.gifProvider = &giphyProvider{}
	plugin.ClientMain(&p)
}
