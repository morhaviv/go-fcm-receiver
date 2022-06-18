package go_fcm_receiver

import "go-fcm-receiver/generic"

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

func (f *FCMClient) subscribeRequest() (string, error) {
	subscribeResponse, err := f.SendSubscribeRequest()
	if err != nil {
		return "", err
	}
	return subscribeResponse.Token, nil
}
