package main

import (
	"fmt"
	"path/filepath"
	"strings"

	manifest "github.com/moussetc/mattermost-plugin-giphy"
	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	provider "github.com/moussetc/mattermost-plugin-giphy/server/internal/provider"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *pluginConf.Configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &pluginConf.Configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
func (p *Plugin) setConfiguration(configuration *pluginConf.Configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(pluginConf.Configuration)
	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "Failed to load plugin configuration")
	}
	p.setConfiguration(configuration)
	if configurationErr := configuration.IsValid(); configurationErr != nil {
		return configurationErr
	}

	rootURL := ""
	if siteURL := p.API.GetConfig().ServiceSettings.SiteURL; siteURL != nil {
		rootURL = strings.TrimSuffix(*siteURL, "/")
	}
	p.rootURL = fmt.Sprintf("%s/plugins/%s", rootURL, manifest.Manifest.Id)

	gifProvider, err := provider.GifProviderGenerator(*configuration, p.errorGenerator, p.rootURL)
	if err != nil {
		return err
	}

	p.gifProvider = gifProvider
	if configuration.DisablePostingWithoutPreview {
		// Force preview
		configuration.CommandTriggerGif = ""
		configuration.CommandTriggerGifWithPreview = triggerGif
	} else {
		// Slack-like syntax
		configuration.CommandTriggerGif = triggerGif
		configuration.CommandTriggerGifWithPreview = triggerGifs
	}

	// Re-register commands since a configuration change can impact the available commands
	return p.RegisterCommands()
}

func (p *Plugin) defineBot() error {
	bot := model.Bot{
		Username:    "gifcommandsplugin",
		DisplayName: manifest.Manifest.Name,
		Description: "Bot for the " + manifest.Manifest.Name + " plugin.",
	}
	botID, ensureBotError := p.pluginClient.Bot.EnsureBot(&bot, pluginapi.ProfileImagePath(filepath.Join("assets", "icon.png")))
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure GIF bot")
	}

	p.botID = botID

	return nil
}
