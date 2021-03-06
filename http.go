package go_fcm_receiver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/golang/protobuf/proto"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type FCMSubscribeResponse struct {
	Token   string `json:"token"`
	PushSet string `json:"pushSet"`
}

func (f *FCMClient) SendCheckInRequest(requestBody *AndroidCheckinRequest) (*AndroidCheckinResponse, error) {
	data, err := proto.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", CheckInUrl, buff)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-protobuf")
	req.Header.Add("User-Agent", "")

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responsePb AndroidCheckinResponse
	err = proto.Unmarshal(result, &responsePb)
	if err != nil {
		return nil, err
	}

	return &responsePb, nil
}

func (f *FCMClient) SendRegisterRequest() (string, error) {
	if f.appId == "" {
		f.CreateAppId()
	}
	values := url.Values{}
	values.Add("app", "org.chromium.linux")
	values.Add("X-subtype", f.appId)
	values.Add("device", strconv.FormatUint(f.AndroidId, 10))
	values.Add("sender", base64.RawURLEncoding.EncodeToString(FcmServerKey))

	req, err := http.NewRequest("POST", RegisterUrl, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "AidLogin "+strconv.FormatUint(f.AndroidId, 10)+":"+strconv.FormatUint(f.SecurityToken, 10))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "")

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	respValues, err := url.ParseQuery(string(result))
	if err != nil {
		return "", err
	}

	if respValues.Get("Error") != "" {
		err = errors.New(respValues.Get("Error"))
		return "", err
	}

	return respValues.Get("token"), nil
}

func (f *FCMClient) SendSubscribeRequest() (*FCMSubscribeResponse, error) {
	publicKey := base64.URLEncoding.EncodeToString(PubBytes(f.publicKey))
	publicKey = strings.ReplaceAll(publicKey, "=", "")
	publicKey = strings.ReplaceAll(publicKey, "+", "")
	publicKey = strings.ReplaceAll(publicKey, "/", "")

	authSecret := base64.RawURLEncoding.EncodeToString(f.authSecret)
	authSecret = strings.ReplaceAll(authSecret, "=", "")
	authSecret = strings.ReplaceAll(authSecret, "+", "")
	authSecret = strings.ReplaceAll(authSecret, "/", "")

	values := url.Values{}
	values.Add("authorized_entity", strconv.FormatInt(f.SenderId, 10))
	values.Add("endpoint", FcmEndpointUrl+"/"+f.GcmToken)
	values.Add("encryption_key", publicKey)
	values.Add("encryption_auth", authSecret)

	req, err := http.NewRequest("POST", FcmSubscribeUrl, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "")

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response FCMSubscribeResponse
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
