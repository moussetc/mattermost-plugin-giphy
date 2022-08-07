package pluginapi

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

type Client struct {
	Bot BotService
}

// System is an interface declaring only the functions from
// mattermost-plugin-api BotService that are used in this plugin
type BotService interface {
	EnsureBot(*model.Bot, ...pluginapi.EnsureBotOption) (retBotID string, retErr error)
}

func NewClient(api plugin.API, driver plugin.Driver) *Client {
	client := pluginapi.NewClient(api, driver)
	return &Client{Bot: &client.Bot}
}
