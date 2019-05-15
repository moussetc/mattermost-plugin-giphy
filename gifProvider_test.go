package main

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestGiphyProviderGetGIFURL(t *testing.T) {
	p := &giphyProvider{}
	config, err := getDefaultConfig(t)
	if err != nil {
		t.Errorf(err.Error())
	}
	url, err := p.getGifURL(config, "cat")
	assert.Nil(t, err)
	assert.NotEmpty(t, url)
}

func getDefaultConfig(t *testing.T) (*GiphyPluginConfiguration, error) {
	yamlFile, err := ioutil.ReadFile("plugin.yaml")
	if err != nil {
		return nil, errors.New("could not open plugin configuration which is necessary for this test")
	}
	var conf *model.Manifest
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, errors.New("could not load plugin configuration which is necessary for this test")
	}

	config := &GiphyPluginConfiguration{
		Language:  "fr",
		Rating:    "",
		Rendition: "fixed_height_small",
		APIKey:    conf.SettingsSchema.Settings[1].Default.(string),
	}

	return config, nil
}
