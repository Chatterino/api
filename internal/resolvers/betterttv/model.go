package betterttv

import "time"

type EmoteAPIUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ProviderID  string `json:"providerId"`
}

type EmoteAPIResponse struct {
	ID             string       `json:"id"`
	Code           string       `json:"code"`
	ImageType      string       `json:"imageType"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
	Global         bool         `json:"global"`
	Live           bool         `json:"live"`
	Sharing        bool         `json:"sharing"`
	ApprovalStatus string       `json:"approvalStatus"`
	User           EmoteAPIUser `json:"user"`
}

type TooltipData struct {
	Code     string
	Type     string
	Uploader string
}
