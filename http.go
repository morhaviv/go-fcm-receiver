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

	if f.AndroidApp == nil || f.InstallationAuthToken == nil {
		values.Add("X-subtype", f.AppId)
		values.Add("app", "org.chromium.linux")
		values.Add("device", strconv.FormatUint(f.AndroidId, 10))
		values.Add("sender", base64.RawURLEncoding.EncodeToString(FcmServerKey))
	} else {
		values.Add("X-subtype", f.AndroidApp.GcmSenderId)
		values.Add("device", strconv.FormatUint(f.AndroidId, 10))
		values.Add("app", f.AndroidApp.AndroidPackage)
		values.Add("cert", f.AndroidApp.AndroidPackageCert)
		values.Add("app_ver", "1")
		values.Add("X-app_ver", "1")
		values.Add("X-osv", "29")
		values.Add("X-cliv", "fiid-21.1.1")
		values.Add("X-gmsv", "220217001")
		values.Add("X-scope", "*")
		values.Add("X-Goog-Firebase-Installations-Auth", *f.InstallationAuthToken)
		values.Add("X-gms_app_id", f.AppId)
		values.Add("X-Firebase-Client", "android-min-sdk/23 fire-core/20.0.0 device-name/a21snnxx device-brand/samsung device-model/a21s android-installer/com.android.vending fire-android/30 fire-installations/17.0.0 fire-fcm/22.0.0 android-platform/ kotlin/1.9.23 android-target-sdk/34")
		values.Add("X-Firebase-Client-Log-Type", "1")
		values.Add("X-app_ver_name", "1")
		values.Add("target_ver", "31")
		values.Add("sender", f.AndroidApp.GcmSenderId)
	}

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
		"fid":         fid,
		"appId":       f.AppId,
		"authVersion": "FIS_v2",
		"sdkVersion":  "w:0.6.4",
	}

	if f.AndroidApp != nil {
		body["sdkVersion"] = "a:17.0.0"
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%sprojects/%s/installations", FirebaseInstallation, f.ProjectID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("x-goog-api-key", f.ApiKey)

	if f.AndroidApp == nil {
		clientInfo := map[string]interface{}{
			"heartbeats": []interface{}{},
			"version":    2,
		}
		clientInfoBytes, err := json.Marshal(clientInfo)
		if err != nil {
			return nil, err
		}
		clientInfoBase64 := base64.StdEncoding.EncodeToString(clientInfoBytes)
		req.Header.Set("x-firebase-client", clientInfoBase64)
	} else {
		req.Header.Set("X-Android-Package", f.AndroidApp.AndroidPackage)
		req.Header.Set("X-Android-Cert", f.AndroidApp.AndroidPackageCert)
		req.Header.Set("x-firebase-client", "android-min-sdk/23 fire-core/20.0.0 device-name/a21snnxx device-brand/samsung device-model/a21s android-installer/com.android.vending fire-android/30 fire-installations/17.0.0 fire-fcm/22.0.0 android-platform/ kotlin/1.9.23 android-target-sdk/34")
		req.Header.Set("x-firebase-client-log-type", "3")
		req.Header.Set("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 11; SM-A217F Build/RP1A.200720.012)")
	}

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
func (f *FCMClient) SendFCMRegisterRequest() (*FCMRegisterResponse, error) {
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
	req.Header.Set("x-goog-firebase-installations-auth", *f.InstallationAuthToken)
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
