package go_fcm_receiver

func registerGCM(appId string) {
	checkInRequest(nil, nil)
}

func checkInRequest(androidId *int64, securityToken *int64) {
	CreateCheckInRequest(androidId, securityToken, "")
}
