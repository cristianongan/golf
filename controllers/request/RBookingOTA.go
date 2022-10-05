package request

type CreateBookingOTABody struct {
	Token        string `json:"Token"`      //
	PlayerName   string `json:"PlayerName"` //
	Contact      string `json:"Contact"`    //
	Note         string `json:"Note"`       // San Golf
	NumBook      int    `json:"NumBook"`
	Holes        int    `json:"Holes"`        // Golf Bag
	IsMainCourse bool   `json:"isMainCourse"` // Số hố
	Tee          string `json:"Tee"`          // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
}
type GetTeeTimeOTAList struct {
	Token        string `form:"Token"`
	CourseCode   string `form:"CourseCode"`
	Date         string `form:"Date"`
	IsMainCourse bool   `form:"isMainCourse"`
	OTA_Code     string `form:"OTA_Code"`
	Guest_Code   string `form:"Guest_Code"`
}
type RTeeTimeStatus struct {
	Token        string `form:"Token"`
	CourseCode   string `form:"CourseCode"`
	DateStr      string `form:"DateStr"`
	Date         string `form:"Date"`
	IsMainCourse bool   `form:"isMainCourse"`
	Tee          string `form:"Tee"`
	TeeOffStr    string `form:"TeeOffStr"`
	Guest_Code   string `form:"Guest_Code"`
	Locktime     int    `form:"Locktime"`
}
