package go_fcm_receiver

import (
	"time"
)

func (f *FCMClient) RegisterGCM() error {
	err := f.checkInRequest()
	if err != nil {
		return err
	}

	err = f.registerRequest()
	if err != nil {
		return err
	}

	return nil
}

func (f *FCMClient) checkInRequest() error {
	androidId := int64(f.AndroidId)
	checkInRequest := CreateCheckInRequest(&androidId, &f.SecurityToken, "")
	responsePb, err := f.SendCheckInRequest(checkInRequest)
	if err != nil {
		return err
	}

	f.AndroidId = *responsePb.AndroidId
	f.SecurityToken = *responsePb.SecurityToken

	return nil
}

func (f *FCMClient) registerRequest() error {
	token, err := f.SendRegisterRequest()
	for err != nil {
		token, err = f.SendRegisterRequest()
		time.Sleep(time.Second)
	}
	f.GcmToken = token
	return nil
}
