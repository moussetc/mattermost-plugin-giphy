package test

import pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"

func MockErrorGenerator() pluginError.PluginError {
	return pluginError.NewPluginErrorGenerator("test")
}
