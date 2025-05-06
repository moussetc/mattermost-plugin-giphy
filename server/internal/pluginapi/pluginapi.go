package pluginapi

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type Client struct {
	Bot BotService
}

// BotService is an interface declaring only the functions from
// mattermost-plugin-api BotService that are used in this plugin
type BotService interface {
	EnsureBot(*model.Bot, ...pluginapi.EnsureBotOption) (retBotID string, retErr error)
}

func NewClient(api plugin.API, driver plugin.Driver) *Client {
	client := pluginapi.NewClient(api, driver)
	return &Client{Bot: &client.Bot}
}
