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
	Token        string `json:"Token"`
	CourseCode   string `json:"CourseCode"`
	Date         string `json:"Date"`
	IsMainCourse bool   `json:"isMainCourse"`
	OTA_Code     string `json:"OTA_Code"`
	Guest_Code   string `json:"Guest_Code"`
}
