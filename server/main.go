package main

import (
	"github.com/mattermost/mattermost-server/v5/plugin"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost-server/v6/plugin"
)

func main() {
	p := Plugin{}
	p.errorGenerator = pluginError.NewPluginErrorGenerator(manifest.Name)
	plugin.ClientMain(&p)
}
