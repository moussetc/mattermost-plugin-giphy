package main

import (
	"net/http"
	"path/filepath"

	. "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	"github.com/moussetc/mattermost-plugin-giphy/server/internal/provider"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *Configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &Configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
func (p *Plugin) setConfiguration(configuration *Configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(Configuration)
	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "Failed to load plugin configuration")
	}

	p.setConfiguration(configuration)

	if configuration.DisplayMode == "" {
		return errors.New("The Display Mode must be configured")
	}

	if configuration.Provider == "" {
		return errors.New("The GIF provider must be configured")
	}
	switch configuration.Provider {
	case "giphy":
		if configuration.APIKey == "" {
			return errors.New("The API Key setting must be set for Giphy")
		}
		p.gifProvider = provider.NewGiphyProvider(http.DefaultClient, p.errorGenerator)
	case "tenor":
		if configuration.APIKey == "" {
			return errors.New("The API Key setting must be set for Tenor")
		}
		p.gifProvider = provider.NewTenorProvider(http.DefaultClient, p.errorGenerator)
	default:
		p.gifProvider = provider.NewGfycatProvider(http.DefaultClient, p.errorGenerator)
	}

	return p.defineBot(configuration.Provider)
}

func (p *Plugin) defineBot(provider string) error {
	bot := model.Bot{
		Username:    "gifcommandsplugin",
		DisplayName: manifest.Name,
		Description: "Bot for the " + manifest.Name + " plugin.",
	}
	botId, ensureBotError := p.Helpers.EnsureBot(&bot, plugin.ProfileImagePath(filepath.Join("assets", "icon.png")))
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure GIF bot.")
	}

	p.botId = botId

	return nil
}
