package go_fcm_receiver

type Checkin struct {
	Type        int         `json:"type"`
	ChromeBuild ChromeBuild `json:"chromeBuild"`
}

type ChromeBuild struct {
	Platform      int    `json:"platform"`
	ChromeVersion string `json:"chromeVersion"`
	Channel       int    `json:"channel"`
}

type CheckInRequest struct {
	UserSerialNumber int     `json:"userSerialNumber"`
	Checkin          Checkin `json:"checkin"`
	Version          int     `json:"version"`
	Id               *int64  `json:"id"`
	SecurityToken    *int64  `json:"securityToken"`
}

func CreateCheckInRequest(androidId *int64, securityToken *int64, chromeVersion string) *CheckInRequest {
	chromeVersion = "63.0.3234.0" // Todo: Delete
	return &CheckInRequest{
		UserSerialNumber: 0,
		Checkin: Checkin{
			Type: 3,
			ChromeBuild: ChromeBuild{
				Platform:      2,
				ChromeVersion: chromeVersion,
				Channel:       1,
			},
		},
		Version:       3,
		Id:            androidId,
		SecurityToken: securityToken,
	}
}
