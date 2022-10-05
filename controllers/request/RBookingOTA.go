package request

import "start/utils"

type CreateBookingOTABody struct {
	Token        string `json:"Token" binding:"required"`      // SHA256(“FLC2020”+ DateStr + TeeOffStr + BookingCode)
	PlayerName   string `json:"PlayerName" binding:"required"` //
	Contact      string `json:"Contact"`                       //
	Note         string `json:"Note"`                          // San Golf
	NumBook      int    `json:"NumBook"`
	Holes        int    `json:"Holes" binding:"required"`   // Số hố
	IsMainCourse bool   `json:"isMainCourse"`               // (true: book vào sân A, false: book vào sân B)
	DateStr      string `json:"DateStr" binding:"required"` // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeeOffStr    string `json:"TeeOffStr" binding:"required"`
	Tee          string `json:"Tee"`
	CourseCode   string `json:"CourseCode" binding:"required"` // uid sân
	// BookAgent    string             `json:"BookAgent"`
	GuestStyle   string             `json:"GuestStyle"`
	AgentCode    string             `json:"AgentCode" binding:"required"`
	CardID       string             `json:"CardID"`
	GreenFee     int64              `json:"GreenFee"`
	CaddieFee    int64              `json:"CaddieFee"`
	BuggyFee     int64              `json:"BuggyFee"`
	Rental       []BookingOTARental `json:"Rental"`
	Caddies      utils.ListString   `json:"Caddies"`
	BookingCode  string             `json:"BookingCode" binding:"required"` // Mã OTA bên VNPay gửi sang để lưu
	EmailConfirm string             `json:"EmailConfirm"`                   // ds email nhận xác nhận booking, cách nhau dấu ";"
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
type RTeeTimeOTA struct {
	Token        string `json:"Token"`
	CourseCode   string `json:"CourseCode"`
	DateStr      string `json:"DateStr"`
	Date         string `json:"Date"`
	IsMainCourse bool   `json:"isMainCourse"`
	Tee          string `json:"Tee"`
	TeeOffStr    string `json:"TeeOffStr"`
	Guest_Code   string `json:"Guest_Code"`
	Locktime     int    `json:"Locktime"`
}
