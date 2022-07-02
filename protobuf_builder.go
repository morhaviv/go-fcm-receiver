package go_fcm_receiver

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"strconv"
)

func CreateLoginRequestRaw(androidId *uint64, securityToken *uint64, persistentIds []string) ([]byte, error) {
	chromeVersion := "chrome-63.0.3234.0"
	domain := "mcs.android.com"

	androidIdFormatted := strconv.FormatUint(*androidId, 10)
	androidIdHex := "android-" + fmt.Sprintf("%x", *androidId)
	securityTokenFormatted := strconv.FormatUint(*securityToken, 10)

	settingName := "new_vc"
	settingValue := "1"
	setting := []*Setting{&Setting{
		Name:  &settingName,
		Value: &settingValue,
	}}
	adaptiveHeartbeat := false
	useRmq2 := true
	authService := LoginRequest_AuthService(2)
	networkType := int32(1)

	if len(persistentIds) > 2 {
		persistentIds = persistentIds[len(persistentIds)-2:]
	}

	req := &LoginRequest{
		Id:                   &chromeVersion,
		Domain:               &domain,
		User:                 &androidIdFormatted,
		Resource:             &androidIdFormatted,
		AuthToken:            &securityTokenFormatted,
		DeviceId:             &androidIdHex,
		LastRmqId:            nil,
		Setting:              setting,
		ReceivedPersistentId: persistentIds,
		AdaptiveHeartbeat:    &adaptiveHeartbeat,
		HeartbeatStat:        nil,
		UseRmq2:              &useRmq2,
		AccountId:            nil,
		AuthService:          &authService,
		NetworkType:          &networkType,
		Status:               nil,
		ClientEvent:          nil,
	}

	loginRequestData, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	return append([]byte{KMCSVersion, KLoginRequestTag, byte(proto.Size(req)), byte(1)}, loginRequestData...), nil
}

func CreateCheckInRequest(androidId *int64, securityToken *uint64) *AndroidCheckinRequest {
	chromeVersion := "63.0.3234.0"
	userSerialNumber := int32(0)
	version := int32(3)
	checkInType := int32(3)
	platform := int32(2)
	channel := int32(1)
	return &AndroidCheckinRequest{
		Id: androidId,
		Checkin: &AndroidCheckinProto{
			Type: (*DeviceType)(&checkInType),
			ChromeBuild: &ChromeBuildProto{
				Platform:      (*ChromeBuildProto_Platform)(&platform),
				ChromeVersion: &chromeVersion,
				Channel:       (*ChromeBuildProto_Channel)(&channel),
			},
		},
		SecurityToken:    securityToken,
		Version:          &version,
		UserSerialNumber: &userSerialNumber,
	}
}

func DecodeHeartbeatPing(data []byte) (*HeartbeatPing, error) {
	var heartbeatPing HeartbeatPing
	err := proto.Unmarshal(data, &heartbeatPing)
	if err != nil {
		return nil, err
	}
	return &heartbeatPing, nil
}

func DecodeHeartbeatAck(data []byte) (*HeartbeatAck, error) {
	var heartbeatAck HeartbeatAck
	err := proto.Unmarshal(data, &heartbeatAck)
	if err != nil {
		return nil, err
	}
	return &heartbeatAck, nil
}

func DecodeLoginRequest(data []byte) (*LoginRequest, error) {
	var loginRequest LoginRequest
	err := proto.Unmarshal(data, &loginRequest)
	if err != nil {
		return nil, err
	}
	return &loginRequest, nil
}

func DecodeLoginResponse(data []byte) (*LoginResponse, error) {
	var loginResponse LoginResponse
	err := proto.Unmarshal(data, &loginResponse)
	if err != nil {
		return nil, err
	}
	return &loginResponse, nil
}

func DecodeClose(data []byte) (*Close, error) {
	var closeObject Close
	err := proto.Unmarshal(data, &closeObject)
	if err != nil {
		return nil, err
	}
	return &closeObject, nil
}

func DecodeIqStanza(data []byte) (*IqStanza, error) {
	var iqStanza IqStanza
	err := proto.Unmarshal(data, &iqStanza)
	if err != nil {
		return nil, err
	}
	return &iqStanza, nil
}

func DecodeDataMessageStanza(data []byte) (*DataMessageStanza, error) {
	var dataMessageStanza DataMessageStanza
	err := proto.Unmarshal(data, &dataMessageStanza)
	if err != nil {
		return nil, err
	}
	return &dataMessageStanza, nil
}

func DecodeStreamErrorStanza(data []byte) (*StreamErrorStanza, error) {
	var streamErrorStanza StreamErrorStanza
	err := proto.Unmarshal(data, &streamErrorStanza)
	if err != nil {
		return nil, err
	}
	return &streamErrorStanza, nil
}
