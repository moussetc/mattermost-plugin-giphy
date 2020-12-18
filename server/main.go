package main

import (
	"github.com/mattermost/mattermost-server/v5/plugin"
	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
)

func main() {
	p := Plugin{}
	p.errorGenerator = pluginError.NewPluginErrorGenerator(manifest.Name)
	plugin.ClientMain(&p)
}
