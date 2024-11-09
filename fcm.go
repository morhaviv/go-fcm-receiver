package go_fcm_receiver

import (
	"encoding/base64"
	"errors"
	"fmt"
)

func (f *FCMClient) registerFCM() error {
	installationToken, err := f.installRequest()
	if err != nil {
		return err
	}
	f.InstallationAuthToken = &installationToken
	if f.AndroidApp == nil {
		token, err := f.registerRequest()
		if err != nil {
			return err
		}
		f.FcmToken = token
	}
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

func (f *FCMClient) installRequest() (string, error) {
	installResponse, err := f.SendFCMInstallRequest()
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to install to the FCM: %s", err.Error()))
		return "", err
	}
	return installResponse.AuthToken.Token, nil
}

func (f *FCMClient) registerRequest() (string, error) {
	registerResponse, err := f.SendFCMRegisterRequest()
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to register to the FCM sender: %s", err.Error()))
		return "", err
	}
	return registerResponse.Token, nil
}
