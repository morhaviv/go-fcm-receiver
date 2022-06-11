package go_fcm_receiver

import (
	"crypto/tls"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	pb "go-fcm-receiver/proto"
	"log"
	"net/http"
)

// FCMClient structure
type FCMClient struct {
	SenderId      int64
	HttpClient    *http.Client
	AppId         string
	GcmToken      string
	FcmToken      string
	Socket        *tls.Conn
	androidId     uint64
	securityToken uint64
	privateKey    string
	publicKey     string
	authSecret    string
	PersistentIds []string
}

func (f *FCMClient) CreateAppId() string {
	f.AppId = fmt.Sprintf(AppIdBase, uuid.New().String())
	return f.AppId
}

func CreateCheckInRequest(androidId *int64, securityToken *uint64, chromeVersion string) *pb.AndroidCheckinRequest {
	// Todo: Do something about this shit
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

type FCMSubscribeResponse struct {
	Token   string `json:"token"`
	PushSet string `json:"pushSet"`
}

func CreateLoginRequestRaw(androidId *uint64, securityToken *uint64, chromeVersion string, persistentIds []string) []byte {
	// Todo: Do something about this shit
	chromeVersion = "chrome-63.0.3234.0" // Todo: Delete
	domain := "mcs.android.com"

	//androidIdFormatted := strconv.FormatUint(*androidId, 10)
	androidIdFormatted := "4630062094884880172"

	//androidIdHex := "android-" + fmt.Sprintf("%x", *androidId)
	androidIdHex := "android-404148f1b59d3f2c"

	//securityTokenFormatted := strconv.FormatUint(*securityToken, 10)
	securityTokenFormatted := "5690696262983213347"

	settingName := "new_vc"
	settingValue := "1"
	setting := []*pb.Setting{&pb.Setting{
		Name:  &settingName,
		Value: &settingValue,
	}}
	adaptiveHeartbeat := false
	useRmq2 := true
	authService := pb.LoginRequest_AuthService(2)
	networkType := int32(1)
	req := &pb.LoginRequest{
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
	return append([]byte{kMCSVersion, kLoginRequestTag, byte(proto.Size(req)), byte(1)}, loginRequestData...)
}
