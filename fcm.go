package go_fcm_receiver

import (
	"encoding/base64"
	"go-fcm-receiver/generic"
	"log"
)

func (f *FCMClient) RegisterFCM() error {
	// Todo: Add option to load old keys
	privateKey, publicKey, authSecret, err := generic.CreateKeys()
	if err != nil {
		return err
	}
	f.privateKey = privateKey
	f.publicKey = publicKey
	f.authSecret = authSecret
	token, err := f.subscribeRequest()
	if err != nil {
		return err
	}
	f.FcmToken = token
	return nil
}

func (f *FCMClient) GetPrivateKeyBase64() (string, error) {
	privateKeyString, err := generic.EncodePrivateKey(f.privateKey)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(privateKeyString), nil
}

func (f *FCMClient) GetAuthSecretBase64() string {
	return base64.StdEncoding.EncodeToString(f.authSecret)
}

func (f *FCMClient) subscribeRequest() (string, error) {
	subscribeResponse, err := f.SendSubscribeRequest()
	if err != nil {
		return "", err
	}
	return subscribeResponse.Token, nil
}
