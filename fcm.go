package go_fcm_receiver

func (f *FCMClient) RegisterFCM() error {
	err := f.checkInRequest()
	if err != nil {
		return err
	}

	return nil
}

func (f *FCMClient) registerFcmRequest() error {
	androidId := int64(f.androidId)
	checkInRequest := CreateCheckInRequest(&androidId, &f.securityToken, "")
	responsePb, err := f.SendCheckInRequest(checkInRequest)
	if err != nil {
		return err
	}

	f.androidId = *responsePb.AndroidId
	f.securityToken = *responsePb.SecurityToken

	return nil
}
