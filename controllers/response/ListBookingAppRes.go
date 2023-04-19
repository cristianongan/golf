package response

import (
	"start/models"
	model_booking "start/models/booking"
)

type Booking struct {
	models.Model
	PartnerUid  string `json:"partner_uid"`  // Hang Golf
	CourseUid   string `json:"course_uid"`   // San Golf
	CourseType  string `json:"course_type"`  // A,B,C
	BookingDate string `json:"booking_date"` // Ex: 06/11/2022

	Bag                 string `json:"bag"`                                            // Golf Bag
	Hole                int    `json:"hole"`                                           // Số hố check in
	HoleBooking         int    `json:"hole_booking"`                                   // Số hố khi booking
	GuestStyle          string `json:"guest_style"`                                    // Guest Style
	GuestStyleName      string `json:"guest_style_name"`                               // Guest Style Name
	CustomerName        string `json:"customer_name"`                                  // Tên khách hàng
	CustomerBookingName string `json:"customer_booking_name" gorm:"type:varchar(256)"` // Tên khách hàng đặt booking

	// MemberCard
	CardId        string `json:"card_id"`         // MembarCard, Card ID cms user nhập vào
	MemberCardUid string `json:"member_card_uid"` // MemberCard Uid, Uid object trong Database

	BagStatus    string `json:"bag_status"`     // Bag status
	CheckInTime  int64  `json:"check_in_time"`  // Time Check In
	CheckOutTime int64  `json:"check_out_time"` // Time Check Out
	TeeType      string `json:"tee_type"`       // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeeTime      string `json:"tee_time"`       // Ex: 16:26 Tee time là thời gian tee off dự kiến

	// Note          string `json:"note" gorm:"type:varchar(500)"`            // Note
	NoteOfBooking string `json:"note_of_booking"` // Note of Booking
	NoteOfGo      string `json:"note_of_go"`      // Note khi trong GO
	LockerNo      string `json:"locker_no"`       // Locker mã số tủ gửi đồ

	// Caddie Id
	CaddieStatus  string                      `json:"caddie_status"` // Caddie status: IN/OUT/INIT
	CaddieBooking string                      `json:"caddie_booking"`
	CaddieId      int64                       `json:"caddie_id"`
	CaddieInfo    model_booking.BookingCaddie `json:"caddie_info,omitempty"` // Caddie Info

	// Buggy Id
	BuggyId   int64                      `json:"buggy_id"`
	BuggyInfo model_booking.BookingBuggy `json:"buggy_info,omitempty"` // Buggy Info

	// Flight Id
	FlightId int64 `json:"flight_id" gorm:"index"`

	// Agency Id
	AgencyId   int64                       `json:"agency_id" gorm:"index"` // Agency
	AgencyInfo model_booking.BookingAgency `json:"agency_info"`

	BookingCode    string `json:"booking_code"`     // cho case tạo nhiều booking có cùng booking code
	BillCode       string `json:"bill_code"`        // hỗ trợ query tính giá
	IsPrivateBuggy *bool  `json:"is_private_buggy"` // Bag có dùng buggy riêng không
	LockBill       *bool  `json:"lock_bill"`        // lễ tân lock bill cho kh để restaurant ko thao tác đc nữa
}
