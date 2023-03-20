package model_booking

import (
	"start/models"
	"start/utils"

	"gorm.io/gorm"
)

/*
 Bỏ ở booking di 1 số trường k cần
 Bỏ CaddieBuggyInOut vì dính khoá ngoại
*/
type BookingDel struct {
	models.Model
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	CourseType  string `json:"course_type" gorm:"type:varchar(100)"`       // A,B,C
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022

	Bag            string `json:"bag" gorm:"type:varchar(100);index"`         // Golf Bag
	Hole           int    `json:"hole"`                                       // Số hố check in
	HoleBooking    int    `json:"hole_booking"`                               // Số hố khi booking
	HoleTimeOut    int    `json:"hole_time_out"`                              // Số hố khi time out
	HoleMoveFlight int    `json:"hole_move_flight"`                           // Số hố trong đã chơi của flight khi bag move sang
	GuestStyle     string `json:"guest_style" gorm:"type:varchar(200);index"` // Guest Style
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(256)"`  // Guest Style Name

	// MemberCard
	CardId        string `json:"card_id" gorm:"index"`                           // MembarCard, Card ID cms user nhập vào
	MemberCardUid string `json:"member_card_uid" gorm:"type:varchar(100);index"` // MemberCard Uid, Uid object trong Database

	// Thêm customer info
	CustomerBookingName  string       `json:"customer_booking_name" gorm:"type:varchar(256)"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string       `json:"customer_booking_phone" gorm:"type:varchar(100)"` // SDT khách hàng đặt booking
	CustomerName         string       `json:"customer_name" gorm:"type:varchar(256)"`          // Tên khách hàng
	CustomerUid          string       `json:"customer_uid" gorm:"type:varchar(256);index"`     // Uid khách hàng
	CustomerType         string       `json:"customer_type" gorm:"type:varchar(256)"`          // Loai khach hang: Member, Guest, Visitor...
	CustomerInfo         CustomerInfo `json:"customer_info,omitempty" gorm:"type:json"`        // Customer Info

	BagStatus         string `json:"bag_status" gorm:"type:varchar(50);index"` // Bag status
	CheckInTime       int64  `json:"check_in_time"`                            // Time Check In
	CheckOutTime      int64  `json:"check_out_time"`                           // Time Check Out
	UndoCheckInTime   int64  `json:"undo_check_in_time"`                       // Time Undo Check In
	CancelBookingTime int64  `json:"cancel_booking_time"`                      // Time cancel booking
	TeeType           string `json:"tee_type" gorm:"type:varchar(50);index"`   // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath           string `json:"tee_path" gorm:"type:varchar(50);index"`   // MORNING, NOON, NIGHT
	TurnTime          string `json:"turn_time" gorm:"type:varchar(30)"`        // Ex: 16:26
	TeeTime           string `json:"tee_time" gorm:"type:varchar(30)"`         // Ex: 16:26 Tee time là thời gian tee off dự kiến
	TeeOffTime        string `json:"tee_off_time" gorm:"type:varchar(30)"`     // Ex: 16:26 Là thời gian thực tế phát bóng
	RowIndex          *int   `json:"row_index"`                                // index trong Flight

	CurrentBagPrice BookingCurrentBagPriceDetail `json:"current_bag_price,omitempty" gorm:"type:json"` // Thông tin phí++: Tính toán lại phí Service items, Tiền cho Subbag
	ListGolfFee     ListBookingGolfFee           `json:"list_golf_fee,omitempty" gorm:"type:json"`     // Thông tin List Golf Fee, Main Bag, Sub Bag
	MushPayInfo     BookingMushPay               `json:"mush_pay_info,omitempty" gorm:"type:json"`     // Mush Pay info
	OtherPaids      utils.ListOtherPaid          `json:"other_paids,omitempty" gorm:"type:json"`       // Other Paids

	// Note          string `json:"note" gorm:"type:varchar(500)"`            // Note
	NoteOfBag     string `json:"note_of_bag" gorm:"type:varchar(500)"`     // Note of Bag
	NoteOfBooking string `json:"note_of_booking" gorm:"type:varchar(500)"` // Note of Booking
	NoteOfGo      string `json:"note_of_go" gorm:"type:varchar(500)"`      // Note khi trong GO
	LockerNo      string `json:"locker_no" gorm:"type:varchar(100)"`       // Locker mã số tủ gửi đồ
	ReportNo      string `json:"report_no" gorm:"type:varchar(200)"`       // Report No
	CancelNote    string `json:"cancel_note" gorm:"type:varchar(300)"`     // Cancel note

	CmsUser    string `json:"cms_user" gorm:"type:varchar(100)"`     // Cms User
	CmsUserLog string `json:"cms_user_log" gorm:"type:varchar(200)"` // Cms User Log

	// Caddie Id
	CaddieStatus  string        `json:"caddie_status" gorm:"type:varchar(50);index"` // Caddie status: IN/OUT/INIT
	CaddieBooking string        `json:"caddie_booking" gorm:"type:varchar(50)"`
	CaddieId      int64         `json:"caddie_id" gorm:"index"`
	CaddieInfo    BookingCaddie `json:"caddie_info,omitempty" gorm:"type:json"` // Caddie Info
	CaddieHoles   int           `json:"caddie_holes"`                           // Lưu lại

	// Buggy Id
	BuggyId   int64        `json:"buggy_id" gorm:"index"`
	BuggyInfo BookingBuggy `json:"buggy_info,omitempty" gorm:"type:json"` // Buggy Info

	// Flight Id
	FlightId int64 `json:"flight_id" gorm:"index"`

	// Agency Id
	AgencyId   int64         `json:"agency_id" gorm:"index"` // Agency
	AgencyInfo BookingAgency `json:"agency_info" gorm:"type:json"`

	// Subs bags
	SubBags utils.ListSubBag `json:"sub_bags,omitempty" gorm:"type:json"` // List Sub Bags

	// Type change hole
	TypeChangeHole string `json:"type_change_hole" gorm:"type:varchar(300)"` // Các loại thay đổi hố

	// Main bags
	MainBags utils.ListSubBag `json:"main_bags,omitempty" gorm:"type:json"` // List Main Bags, thêm main bag sẽ thanh toán những cái gì
	// Main bug for Pay: Mặc định thanh toán all, Nếu có trong list này thì k thanh toán
	MainBagPay utils.ListString `json:"main_bag_pay,omitempty" gorm:"type:json"` // Main Bag không thanh toán những phần này ở sub bag này
	SubBagNote string           `json:"sub_bag_note" gorm:"type:varchar(500)"`   // Note of SubBag

	InitType string `json:"init_type" gorm:"type:varchar(50);index"` // BOOKING: Tạo booking xong checkin, CHECKIN: Check In xong tạo Booking luôn

	// CaddieBuggyInOut   []CaddieBuggyInOut      `json:"caddie_buggy_in_out"`
	BookingCode        string                  `json:"booking_code" gorm:"type:varchar(100);index"`         // cho case tạo nhiều booking có cùng booking code
	BookingCodePartner string                  `json:"booking_code_partner" gorm:"type:varchar(100);index"` // Booking code của partner
	BookingRestaurant  utils.BookingRestaurant `json:"booking_restaurant,omitempty" gorm:"type:json"`
	BookingRetal       utils.BookingRental     `json:"booking_retal,omitempty" gorm:"type:json"`
	BookingSourceId    string                  `json:"booking_source_id" gorm:"type:varchar(50);index"`

	MemberUidOfGuest  string `json:"member_uid_of_guest" gorm:"type:varchar(50);index"` // Member của Guest đến chơi cùng
	MemberNameOfGuest string `json:"member_name_of_guest" gorm:"type:varchar(200)"`     // Member của Guest đến chơi cùng

	HasBookCaddie bool   `json:"has_book_caddie" gorm:"default:0"`
	TimeOutFlight int64  `json:"time_out_flight,omitempty"`                // Thời gian out Flight
	BillCode      string `json:"bill_code" gorm:"type:varchar(100);index"` // hỗ trợ query tính giá
	SeparatePrice bool   `json:"separate_price" gorm:"default:0"`          // Giá riêng

	ShowCaddieBuggy   *bool                                `json:"show_caddie_buggy" gorm:"default:1"` // Sau add round thì không hiển thị caddie buggy
	IsPrivateBuggy    *bool                                `json:"is_private_buggy" gorm:"default:0"`  // Bag có dùng buggy riêng không
	MovedFlight       *bool                                `json:"moved_flight" gorm:"default:0"`      // Đánh dấu booking đã move flight chưa
	AddedRound        *bool                                `json:"added_flight" gorm:"default:0"`      // Đánh dấu booking đã add chưa
	AgencyPaid        utils.ListBookingAgencyPayForBagData `json:"agency_paid,omitempty" gorm:"type:json"`
	AgencyPrePaid     utils.ListBookingAgencyPayForBagData `json:"agency_pre_paid,omitempty" gorm:"type:json"`  // Tiền Agency trả trước
	LockBill          *bool                                `json:"lock_bill" gorm:"default:0"`                  // lễ tân lock bill cho kh để restaurant ko thao tác đc nữa
	AgencyPaidAll     *bool                                `json:"agency_paid_all" gorm:"default:0"`            // Đánh dấu agency trả all fee cho kh
	LastBookingStatus string                               `json:"last_booking_status" gorm:"type:varchar(50)"` // Đánh dấu trạng thái cuối cùng của booking
}

func (item *BookingDel) Create(db *gorm.DB) error {
	return db.Create(item).Error
}
