package configuration

// Configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
type Configuration struct {
	Provider                     string
	DisplayMode                  string
	Rating                       string
	Language                     string
	Rendition                    string
	RenditionGfycat              string
	RenditionTenor               string
	APIKey                       string
	DisablePostingWithoutPreview bool
	RandomSearch                 bool
	// Computed fields:
	CommandTriggerGif            string
	CommandTriggerGifWithPreview string
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *Configuration) Clone() *Configuration {
	var clone = *c
	return &clone
}

const (
	// DisplayModeEmbedded display GIFs as Markdown embedded images
	DisplayModeEmbedded = "embedded"
	// DisplayModeFullURL displays GIFs as raw URLs using image preview
	DisplayModeFullURL = "full_url"
)
