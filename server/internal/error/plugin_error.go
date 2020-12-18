package error

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
)

// PluginError create appErrors enriched with plugin name for better logging
type PluginError interface {
	FromError(message string, err error) *model.AppError
	FromMessage(message string) *model.AppError
}

// NewPluginErrorGenerator returns the default PluginError
func NewPluginErrorGenerator(manifestName string) PluginError {
	return &pluginError{where: manifestName}
}

type pluginError struct {
	where string
}

//appError generates a normalized error for this plugin
func (e *pluginError) FromError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError(e.where, message, nil, errorMessage, http.StatusBadRequest)
}

//appError generates a normalized error for this plugin
func (e *pluginError) FromMessage(message string) *model.AppError {
	return e.FromError(message, nil)
}
