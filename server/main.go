package main

import (
	manifest "github.com/moussetc/mattermost-plugin-giphy"

	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

	"github.com/mattermost/mattermost/server/public/plugin"
)

func main() {
	p := Plugin{}
	p.errorGenerator = pluginError.NewPluginErrorGenerator(manifest.Manifest.Name)
	plugin.ClientMain(&p)
}
