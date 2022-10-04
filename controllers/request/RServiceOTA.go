package request

type ServiceGolfDataBody struct {
	Token      string `json:"Token"`
	CourseCode string `json:"CourseCode"`
}

type CheckServiceGolfBody struct {
	Token      string `json:"Token"`
	CourseCode string `json:"CourseCode"`
	RenTalCode string `json:"RenTalCode"`
	CaddieNo   int64  `json:"CaddieNo"`
	DateStr    string `json:"DateStr"`
	TeeOffStr  string `json:"TeeOffStr"`
	Qty        int64  `json:"Qty"`
}
