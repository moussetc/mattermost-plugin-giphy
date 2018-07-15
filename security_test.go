package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	encryptionKey string = "kjhflkshdlfjkqghkjdghlqkjdfhglqk"
)

func TestEncrypt(t *testing.T) {
	value, err := encryptParameters(encryptionKey, "userID", "channelID", "gifURL", "keyword")
	assert.Nil(t, err)
	assert.NotEqual(t, " ", value)
	fmt.Println(value)

	userID, channelID, gifURL, keyword, err := decryptParameters(encryptionKey, value)
	assert.Nil(t, err)
	assert.Equal(t, "userID", userID)
	assert.Equal(t, "channelID", channelID)
	assert.Equal(t, "gifURL", gifURL)
	assert.Equal(t, "keyword", keyword)
}
