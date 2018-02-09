package main

import (
	"io/ioutil"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestGiphyProviderGetGIFURL(t *testing.T) {
	yamlFile, err := ioutil.ReadFile("plugin.yaml")
	if err != nil {
		t.Errorf("Could not open plugin configuration which is necessary for this test.")
		return
	}
	var conf *model.Manifest
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		t.Errorf("Could not load plugin configuration which is necessary for this test.")
		return
	}

	p := &giphyProvider{}
	config := &GiphyPluginConfiguration{
		Language:  "fr",
		Rating:    "",
		Rendition: "fixed_height_small",
		APIKey:    conf.SettingsSchema.Settings[0].Default.(string),
	}
	url, err := p.getGifURL(config, "cat")
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
}
