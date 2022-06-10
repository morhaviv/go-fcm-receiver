package go_fcm_receiver

import (
	"fmt"
	"strconv"
)

func (f *FCMClient) RegisterGCM() {
	f.checkInRequest(nil, nil)
}

func (f *FCMClient) checkInRequest(androidId *int64, securityToken *uint64) error {
	checkInRequest := CreateCheckInRequest(androidId, securityToken, "")
	responsePb, err := f.SendCheckInRequest(checkInRequest)
	if err != nil {
		return err
	}
	fmt.Println(responsePb)
	fmt.Println(*responsePb.AndroidId)
	fmt.Println(*responsePb.SecurityToken)
	fmt.Println(*responsePb.VersionInfo)
	fmt.Println(string(strconv.FormatUint(responsePb.GetAndroidId(), 10)))
	return nil
}
