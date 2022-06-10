package go_fcm_receiver

import "net/http"
import pb "go-fcm-receiver/proto"

// FCMClient structure
type FCMClient struct {
	SenderId   int64
	HttpClient *http.Client
}

func CreateCheckInRequest(androidId *int64, securityToken *uint64, chromeVersion string) *pb.AndroidCheckinRequest {
	chromeVersion = "63.0.3234.0" // Todo: Delete
	userSerialNumber := int32(0)
	version := int32(3)
	chekinType := int32(3)
	platform := int32(2)
	channel := int32(1)
	return &pb.AndroidCheckinRequest{
		Id: androidId,
		Checkin: &pb.AndroidCheckinProto{
			Type: (*pb.DeviceType)(&chekinType),
			ChromeBuild: &pb.ChromeBuildProto{
				Platform:      (*pb.ChromeBuildProto_Platform)(&platform),
				ChromeVersion: &chromeVersion,
				Channel:       (*pb.ChromeBuildProto_Channel)(&channel),
			},
		},
		SecurityToken:    securityToken,
		Version:          &version,
		UserSerialNumber: &userSerialNumber,
	}

}
