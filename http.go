package go_fcm_receiver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

type AuthToken struct {
	Token string `json:"token"`
}

type FCMInstallationResponse struct {
	AuthToken AuthToken `json:"authToken"`
}

type FCMRegisterResponse struct {
	Token   string `json:"token"`
	PushSet string `json:"pushSet"`
}

// SendGCMCheckInRequest GCM Checkin Request
func (f *FCMClient) SendGCMCheckInRequest(requestBody *AndroidCheckinRequest) (*AndroidCheckinResponse, error) {
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

// SendGCMRegisterRequest GCM Register Request
func (f *FCMClient) SendGCMRegisterRequest() (string, error) {
	values := url.Values{}
	values.Add("app", "org.chromium.linux")
	values.Add("X-subtype", f.AppId)
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

// SendFCMInstallRequest FCM Installation Request
func (f *FCMClient) SendFCMInstallRequest() (*FCMInstallationResponse, error) {
	fid, err := GenerateFirebaseFID()
	if err != nil {
		return nil, err
	}

	body := map[string]string{
		"appId":       f.AppId,
		"authVersion": "FIS_v2",
		"fid":         fid,
		"sdkVersion":  "w:0.6.4",
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	clientInfo := map[string]interface{}{
		"heartbeats": []interface{}{},
		"version":    2,
	}
	clientInfoBytes, err := json.Marshal(clientInfo)
	if err != nil {
		return nil, err
	}
	clientInfoBase64 := base64.StdEncoding.EncodeToString(clientInfoBytes)

	req, err := http.NewRequest("POST", fmt.Sprintf("%sprojects/%s/installations", FirebaseInstallation, f.ProjectID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-firebase-client", clientInfoBase64)
	req.Header.Set("x-goog-api-key", f.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response FCMInstallationResponse
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SendFCMRegisterRequest FCM Registration Request
func (f *FCMClient) SendFCMRegisterRequest(installationAuthToken string) (*FCMRegisterResponse, error) {
	publicKey := base64.URLEncoding.EncodeToString(PubBytes(f.publicKey))
	publicKey = strings.ReplaceAll(publicKey, "=", "")
	publicKey = strings.ReplaceAll(publicKey, "+", "")
	publicKey = strings.ReplaceAll(publicKey, "/", "")

	authSecret := base64.RawURLEncoding.EncodeToString(f.authSecret)
	authSecret = strings.ReplaceAll(authSecret, "=", "")
	authSecret = strings.ReplaceAll(authSecret, "+", "")
	authSecret = strings.ReplaceAll(authSecret, "/", "")

	body := map[string]interface{}{
		"web": map[string]string{
			"applicationPubKey": f.VapidKey,
			"auth":              authSecret,
			"endpoint":          fmt.Sprintf("%s/%s", FcmEndpointUrl, f.GcmToken),
			"p256dh":            publicKey,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%sprojects/%s/registrations", FirebaseRegistrationUrl, f.ProjectID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-goog-api-key", f.ApiKey)
	req.Header.Set("x-goog-firebase-installations-auth", installationAuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response FCMRegisterResponse
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
