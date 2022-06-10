package go_fcm_receiver

import "fmt"

func (f *FCMClient) RegisterFCM() error {
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
	fmt.Println(f.GcmToken)
	fmt.Println(f.FcmToken)

	return nil
}

func (f *FCMClient) subscribeRequest() (string, error) {
	subscribeResponse, err := f.SendSubscribeRequest()
	if err != nil {
		return "", err
	}
	return subscribeResponse.Token, nil
}
