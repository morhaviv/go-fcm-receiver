package go_fcm_receiver

import (
	"errors"
	"fmt"
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
	checkInRequest := CreateCheckInRequest(&androidId, &f.SecurityToken)
	responsePb, err := f.SendCheckInRequest(checkInRequest)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to send GCM checkIn request: %s", err.Error()))
		return err
	}

	f.AndroidId = *responsePb.AndroidId
	f.SecurityToken = *responsePb.SecurityToken

	return nil
}

func (f *FCMClient) registerRequest() error {
	// Server sometimes returns an error `(Error=PHONE_REGISTRATION_ERROR)` for no reason, so we're trying multiple times
	token, err := f.SendRegisterRequest()
	i := 0
	for err != nil && i < 10 {
		i += 1
		time.Sleep(time.Second)
		token, err = f.SendRegisterRequest()
	}
	if i > 10 {
		err = errors.New(fmt.Sprintf("failed to send GCM register request: %s", err.Error()))
		return err
	}
	f.GcmToken = token
	return nil
}
