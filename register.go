package go_fcm_receiver

func (f *FCMClient) Register() error {
	err := f.RegisterGCM()
	if err != nil {
		return err
	}
	err = f.RegisterFCM()
	if err != nil {
		return err
	}

	return nil
}
