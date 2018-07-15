package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
)

func encryptParameters(key string, userID string, channelID string, gifURL string, keyword string) (url.Values, error) {
	secureParams := url.Values{}
	secureParams.Add("channelId", channelID)
	secureParams.Add("userId", userID)
	secureParams.Add("keyword", keyword)
	secureParams.Add("gifUrl", gifURL)

	encryptedParams, err := encrypt([]byte(key), secureParams.Encode())
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("data", string(encryptedParams))
	return params, nil
}

func decryptParameters(key string, data url.Values) (userID string, channelID string, gifURL string, keyword string, err error) {
	query, err := decrypt([]byte(key), data.Get("data"))
	if err != nil {
		return "", "", "", "", err
	}
	values, err := url.ParseQuery(string(query))
	if err != nil {
		return "", "", "", "", err
	}

	userID = values.Get("userId")
	channelID = values.Get("channelId")
	gifURL = values.Get("gifUrl")
	keyword = values.Get("keyword")
	return
}

func encrypt(key []byte, message string) (string, error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", appError("failed to initiate new cipher", err)
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", appError("failed to generate iv", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func decrypt(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("ciphertext block size is too short")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}
