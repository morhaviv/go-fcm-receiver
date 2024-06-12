package go_fcm_receiver

import (
	"bytes"
	"crypto/rand"
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

// SendCheckInRequest GCM Checkin Request
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

// SendRegisterRequest GCM Register Request
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

// SendSubscribeRequest FCM Deprecated subscribe request
//func (f *FCMClient) SendSubscribeRequest() (*FCMSubscribeResponse, error) {
//	publicKey := base64.URLEncoding.EncodeToString(PubBytes(f.publicKey))
//	publicKey = strings.ReplaceAll(publicKey, "=", "")
//	publicKey = strings.ReplaceAll(publicKey, "+", "")
//	publicKey = strings.ReplaceAll(publicKey, "/", "")
//
//	authSecret := base64.RawURLEncoding.EncodeToString(f.authSecret)
//	authSecret = strings.ReplaceAll(authSecret, "=", "")
//	authSecret = strings.ReplaceAll(authSecret, "+", "")
//	authSecret = strings.ReplaceAll(authSecret, "/", "")
//
//	values := url.Values{}
//	values.Add("authorized_entity", strconv.FormatInt(f.SenderId, 10))
//	values.Add("endpoint", FcmEndpointUrl+"/"+f.GcmToken)
//	values.Add("encryption_key", publicKey)
//	values.Add("encryption_auth", authSecret)
//
//	req, err := http.NewRequest("POST", FcmSubscribeUrl, strings.NewReader(values.Encode()))
//	if err != nil {
//		return nil, err
//	}
//
//	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Add("User-Agent", "")
//
//	resp, err := f.HttpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	result, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return nil, err
//	}
//
//	var response FCMSubscribeResponse
//	err = json.Unmarshal(result, &response)
//	if err != nil {
//		return nil, err
//	}
//
//	return &response, nil
//}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	FIREBASE_INSTALLATION = "https://firebaseinstallations.googleapis.com/v1/"
	FCM_REGISTRATION      = "https://fcmregistrations.googleapis.com/v1/"
	FCM_ENDPOINT          = "https://fcm.googleapis.com/fcm/send"
)

func generateFirebaseFID() (string, error) {
	// A valid FID has exactly 22 base64 characters, which is 132 bits, or 16.5
	// bytes. Our implementation generates a 17 byte array instead.
	fid := make([]byte, 17)
	_, err := rand.Read(fid)
	if err != nil {
		return "", err
	}

	// Replace the first 4 random bits with the constant FID header of 0b0111.
	fid[0] = 0b01110000 + (fid[0] % 0b00010000)

	return base64.StdEncoding.EncodeToString(fid), nil
}

func (f *FCMClient) SendFCMInstallRequest() (*FCMInstallationResponse, error) {
	fid, err := generateFirebaseFID()
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

	fmt.Println(string(bodyBytes))

	req, err := http.NewRequest("POST", fmt.Sprintf("%sprojects/%s/installations", FIREBASE_INSTALLATION, f.ProjectID), bytes.NewBuffer(bodyBytes))
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

	fmt.Println(result)
	fmt.Println(string(result))

	var response FCMInstallationResponse
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

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
			"endpoint":          fmt.Sprintf("%s/%s", FCM_ENDPOINT, f.GcmToken),
			"p256dh":            publicKey,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%sprojects/%s/registrations", FCM_REGISTRATION, f.ProjectID), bytes.NewBuffer(bodyBytes))
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
