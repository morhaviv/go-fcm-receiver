package proto

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"go-fcm-receiver"
	"log"
	"strconv"
)

func CreateLoginRequestRaw(androidId *uint64, securityToken *uint64, chromeVersion string, persistentIds []string) []byte {
	// Todo: Consider moving to proto/builder.go
	// Todo: Do something about this shit
	chromeVersion = "chrome-63.0.3234.0" // Todo: Delete
	domain := "mcs.android.com"

	androidIdFormatted := strconv.FormatUint(*androidId, 10)
	//androidIdFormatted := "4630062094884880172"

	androidIdHex := "android-" + fmt.Sprintf("%x", *androidId)
	//androidIdHex := "android-404148f1b59d3f2c"

	securityTokenFormatted := strconv.FormatUint(*securityToken, 10)
	//securityTokenFormatted := "5690696262983213347"

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
		log.Print(err)
		return nil
	}
	return append([]byte{go_fcm_receiver.KMCSVersion, go_fcm_receiver.KLoginRequestTag, byte(proto.Size(req)), byte(1)}, loginRequestData...)
}

func CreateCheckInRequest(androidId *int64, securityToken *uint64, chromeVersion string) *AndroidCheckinRequest {
	// Todo: Consider moving to proto/builder.go
	// Todo: Do something about this shit
	chromeVersion = "63.0.3234.0" // Todo: Delete
	userSerialNumber := int32(0)
	version := int32(3)
	chekinType := int32(3)
	platform := int32(2)
	channel := int32(1)
	return &AndroidCheckinRequest{
		Id: androidId,
		Checkin: &AndroidCheckinProto{
			Type: (*DeviceType)(&chekinType),
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
