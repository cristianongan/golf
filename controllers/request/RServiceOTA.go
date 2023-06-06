package request

type ServiceGolfDataBody struct {
	Token      string `json:"Token"`
	CourseCode string `json:"CourseCode"`
}

type CheckServiceGolfBody struct {
	Token      string `json:"Token"`
	CourseCode string `json:"CourseCode"`
	RenTalCode string `json:"RenTalCode"`
	CaddieNo   string `json:"CaddieNo"`
	DateStr    string `json:"DateStr"`
	TeeOffStr  string `json:"TeeOffStr"`
	Qty        int64  `json:"Qty"`
}

type GetListFeeServiceOTABody struct {
	Token      string `json:"Token"`
	CourseCode string `json:"CourseCode"`
	DateStr    string `json:"DateStr"`
	OTACode    string `json:"OTACode"`
	Hole       int    `json:"Hole"`
}
type CheckMemberCardBody struct {
	Token      string `json:"Token"`
	CourseCode string `json:"CourseCode"`
	CardId     string `json:"CardId"`
	OtaCode    string `json:"OTA_Code"`
}

type CheckCaddieBody struct {
	Token      string `json:"Token"`
	CourseCode string `json:"CourseCode"`
	CaddieCode string `json:"CaddieCode"`
	OtaCode    string `json:"OTA_Code"`
}
