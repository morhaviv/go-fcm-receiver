package go_fcm_receiver

import (
	"encoding/base64"
	"errors"
	"fmt"
)

func (f *FCMClient) registerFCM() error {
	privateKey, publicKey, authSecret, err := CreateKeys()
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
	privateKeyString, err := EncodePrivateKey(f.privateKey)
	if err != nil {
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
		err = errors.New(fmt.Sprintf("failed to subscribe to the FCM sender: %s", err.Error()))
		return "", err
	}
	return subscribeResponse.Token, nil
}
