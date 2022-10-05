package request

import "start/utils"

type CreateBookingOTABody struct {
	Token        string             `json:"Token"`      // SHA256(“FLC2020”+ DateStr + TeeOffStr + BookingCode)
	PlayerName   string             `json:"PlayerName"` //
	Contact      string             `json:"Contact"`    //
	Note         string             `json:"Note"`       // San Golf
	NumBook      int                `json:"NumBook"`
	Holes        int                `json:"Holes"`        // Số hố
	IsMainCourse bool               `json:"isMainCourse"` // (true: book vào sân A, false: book vào sân B)
	DateStr      string             `json:"DateStr"`      // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeeOffStr    string             `json:"TeeOffStr"`
	CourseCode   string             `json:"CourseCode"`
	BookAgent    string             `json:"BookAgent"`
	GuestStyle   string             `json:"GuestStyle"`
	AgentCode    string             `json:"AgentCode"`
	CardID       string             `json:"CardID"`
	GreenFee     int64              `json:"GreenFee"`
	CaddieFee    int64              `json:"CaddieFee"`
	BuggyFee     int64              `json:"BuggyFee"`
	Rental       []BookingOTARental `json:"Rental"`
	Caddies      utils.ListString   `json:"Caddies"`
	BookingCode  string             `json:"BookingCode"`  // Mã OTA bên VNPay gửi sang để lưu
	EmailConfirm string             `json:"EmailConfirm"` // ds email nhận xác nhận booking, cách nhau dấu ";"
}

type BookingOTARental struct {
	QTy   int    `json:"QTy"`
	Code  string `json:"Code"`
	Price int64  `json:"Price"`
}

type GetTeeTimeOTAList struct {
	Token        string `json:"Token"`
	CourseCode   string `json:"CourseCode"`
	Date         string `json:"Date"`
	IsMainCourse bool   `json:"isMainCourse"`
	OTA_Code     string `json:"OTA_Code"`
	Guest_Code   string `json:"Guest_Code"`
}
