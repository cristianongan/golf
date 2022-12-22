package model_booking

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Booking
// omitempty: xứ lý khi các field trả về rỗng
type Booking struct {
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
	CaddieStatus string        `json:"caddie_status" gorm:"type:varchar(50);index"` // Caddie status: IN/OUT/INIT
	CaddieId     int64         `json:"caddie_id" gorm:"index"`
	CaddieInfo   BookingCaddie `json:"caddie_info,omitempty" gorm:"type:json"` // Caddie Info
	CaddieHoles  int           `json:"caddie_holes"`                           // Lưu lại

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

	CaddieBuggyInOut   []CaddieBuggyInOut      `json:"caddie_buggy_in_out" gorm:"foreignKey:BookingUid;references:Uid"`
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

	ListServiceItems []BookingServiceItem                 `json:"list_service_items,omitempty" gorm:"-:migration"` // List service item: rental, proshop, restaurant, kiosk
	ShowCaddieBuggy  *bool                                `json:"show_caddie_buggy" gorm:"default:1"`              // Sau add round thì không hiển thị caddie buggy
	IsPrivateBuggy   *bool                                `json:"is_private_buggy" gorm:"default:0"`               // Bag có dùng buggy riêng không
	MovedFlight      *bool                                `json:"moved_flight" gorm:"default:0"`                   // Đánh dấu booking đã move flight chưa
	AddedRound       *bool                                `json:"added_flight" gorm:"default:0"`                   // Đánh dấu booking đã add chưa
	AgencyPaid       utils.ListBookingAgencyPayForBagData `json:"agency_paid,omitempty" gorm:"type:json"`
	LockBill         *bool                                `json:"lock_bill" gorm:"default:0"` // lễ tân lock bill cho kh để restaurant ko thao tác đc nữa
}

type FlyInfoResponse struct {
	Booking
	TeeFlight       int    `json:"tee_flight,omitempty" gorm:"-:migration"`
	TeeOffFlight    string `json:"tee_off_flight,omitempty" gorm:"-:migration"`
	TurnFlight      string `json:"turn_flight,omitempty" gorm:"-:migration"`
	GroupNameFlight string `json:"group_name_flight,omitempty" gorm:"-:migration"`
}

type BagDetail struct {
	Booking
	Rounds models.ListRound `json:"rounds"`
	//ListServiceItems ListBookingServiceItems `json:"list_service_items,omitempty"`
}

type GolfFeeOfBag struct {
	Booking
	ListRoundOfSubBag []RoundOfBag `json:"list_round_of_sub_bag"`
}

type PaymentOfBag struct {
	BagDetail
	ListRoundOfSubBag []RoundOfBag `json:"list_round_of_sub_bag"`
}

type RoundOfBag struct {
	Bag         string           `json:"bag"`
	BookingCode string           `json:"booking_code"`
	PlayerName  string           `json:"player_name"`
	Rounds      models.ListRound `json:"rounds"`
}

type BookingForListServiceIems struct {
	PartnerUid       string                  `json:"partner_uid"`                                                              // Hang Golf
	CourseUid        string                  `json:"course_uid"`                                                               // San Golf
	BookingDate      string                  `json:"booking_date"`                                                             // Ex: 06/11/2022
	Bag              string                  `json:"bag"`                                                                      // Golf Bag
	ListServiceItems ListBookingServiceItems `json:"list_service_items,omitempty" gorm:"foreignKey:BookingUid;references:Uid"` // List service item: rental, proshop, restaurant, kiosk
	CheckInTime      int64                   `json:"check_in_time"`
	CustomerName     string                  `json:"customer_name"`
}
type GetListBookingWithListServiceItems struct {
	PartnerUid  string
	CourseUid   string
	FromDate    string
	ToDate      string
	GolfBag     string
	PlayerName  string
	ServiceType string
}

type BookingForReportMainBagSubBags struct {
	models.Model
	PartnerUid string `json:"partner_uid"` // Hang Golf
	CourseUid  string `json:"course_uid"`  // San Golf

	BookingDate string `json:"booking_date"` // Ex: 06/11/2022

	Bag            string `json:"bag"`              // Golf Bag
	Hole           int    `json:"hole"`             // Số hố
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name

	CheckOutTime int64  `json:"check_out_time"` // Time Check Out
	BagStatus    string `json:"bag_status"`     // Bag status

	// Subs bags
	SubBags utils.ListSubBag `json:"sub_bags,omitempty"` // List Sub Bags

	// Main bags
	MainBags utils.ListSubBag `json:"main_bags,omitempty"` // List Main Bags, thêm main bag sẽ thanh toán những cái gì

	MushPayInfo     BookingMushPay               `json:"mush_pay_info,omitempty"` // Mush Pay info
	CurrentBagPrice BookingCurrentBagPriceDetail `json:"current_bag_price,omitempty"`
}

type CaddieBuggyInOut CaddieBuggyInOutNoteForBooking

type CaddieBuggyInOutNoteForBooking struct {
	models.ModelId
	PartnerUid     string `json:"partner_uid"`
	CourseUid      string `json:"course_uid"`
	BookingUid     string `json:"booking_uid"`
	CaddieId       int64  `json:"caddie_id"`
	CaddieCode     string `json:"caddie_code"`
	BuggyId        int64  `json:"buggy_id"`
	BuggyCode      string `json:"buggy_code"`
	Note           string `json:"note"`
	CaddieType     string `json:"caddie_type"`
	BuggyType      string `json:"buggy_type"`
	Hole           int    `json:"hole"`
	BagShareBuggy  string `json:"bag_share_buggy"`
	IsPrivateBuggy *bool  `json:"is_private_buggy"`
}

type BuggyInOut BuggyInOutNoteForBooking

type BuggyInOutNoteForBooking struct {
	models.ModelId
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	BookingUid string `json:"booking_uid"`
	BuggyId    int64  `json:"buggy_id"`
	BuggyCode  string `json:"buggy_code"`
	Note       string `json:"note"`
	Type       string `json:"type"`
}

type BookingForFlightRes struct {
	models.Model
	PartnerUid string `json:"partner_uid"` // Hang Golf
	CourseUid  string `json:"course_uid"`  // San Golf

	BookingDate string `json:"booking_date"` // Ex: 06/11/2022

	Bag            string `json:"bag"`              // Golf Bag
	Hole           int    `json:"hole"`             // Số hố
	GuestStyle     string `json:"guest_style" `     // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name

	CustomerName string `json:"customer_name"` // Tên khách hàng

	BagStatus    string `json:"bag_status" gorm:"type:varchar(50);index"` // Check In Out status
	CheckInTime  int64  `json:"check_in_time"`                            // Time Check In
	CheckOutTime int64  `json:"check_out_time"`                           // Time Check Out
	TeeType      string `json:"tee_type" gorm:"type:varchar(50);index"`   // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath      string `json:"tee_path" gorm:"type:varchar(50);index"`   // MORNING, NOON, NIGHT
	TurnTime     string `json:"turn_time" gorm:"type:varchar(30)"`        // Ex: 16:26
	TeeTime      string `json:"tee_time" gorm:"type:varchar(30)"`         // Ex: 16:26 Tee time là thời gian tee off dự kiến
	TeeOffTime   string `json:"tee_off_time" gorm:"type:varchar(30)"`     // Ex: 16:26 Là thời gian thực tế phát bóng
	RowIndex     int    `json:"row_index"`                                // index trong Flight

	// Caddie Id
	CaddieStatus string        `json:"caddie_status" ` // Caddie status: IN/OUT/INIT
	CaddieId     int64         `json:"caddie_id" `
	CaddieInfo   BookingCaddie `json:"caddie_info,omitempty" ` // Caddie Info
	CaddieHoles  int           `json:"caddie_holes"`           // Lưu lại

	// Buggy Id
	BuggyId   int64        `json:"buggy_id" `
	BuggyInfo BookingBuggy `json:"buggy_info,omitempty" ` // Buggy Info

	// Flight Id
	FlightId int64 `json:"flight_id" `

	// Agency Id
	AgencyId   int64         `json:"agency_id" `
	AgencyInfo BookingAgency `json:"agency_info" `
}

type BookingForSubBag struct {
	models.Model
	PartnerUid string `json:"partner_uid" ` // Hang Golf
	CourseUid  string `json:"course_uid" `  // San Golf

	BookingDate string `json:"booking_date" ` // Ex: 06/11/2022

	Bag            string `json:"bag" `              // Golf Bag
	Hole           int    `json:"hole"`              // Số hố
	GuestStyle     string `json:"guest_style" `      // Guest Style
	GuestStyleName string `json:"guest_style_name" ` // Guest Style Name

	CustomerName string `json:"customer_name" ` // Tên khách hàng
	// Subs bags
	SubBags utils.ListSubBag `json:"sub_bags,omitempty" ` // List Sub Bags

	// Main bags
	MainBags utils.ListSubBag `json:"main_bags,omitempty" ` // List Main Bags, thêm main bag sẽ thanh toán những cái gì
}

type CustomerInfo struct {
	Uid         string `json:"uid"`
	PartnerUid  string `json:"partner_uid"`  // Hang Golf
	CourseUid   string `json:"course_uid"`   // San Golf
	Type        string `json:"type"`         // Loai khach hang: Member, Guest, Visitor...
	Name        string `json:"name"`         // Ten KH
	Dob         int64  `json:"dob"`          // Ngay sinh
	Sex         int    `json:"sex"`          // giới tính
	Avatar      string `json:"avatar"`       // ảnh đại diện
	Nationality string `json:"nationality"`  // Quốc gia
	Phone       string `json:"phone"`        // So dien thoai
	CellPhone   string `json:"cell_phone"`   // So dien thoai
	Fax         string `json:"fax"`          // So Fax
	Email       string `json:"email"`        // Email
	Address1    string `json:"address1"`     // Dia chi
	Address2    string `json:"address2"`     // Dia chi
	Job         string `json:"job"`          // Ex: Ngan hang
	Position    string `json:"position"`     // Ex: Chu tich
	CompanyName string `json:"company_name"` // Ten cong ty
	CompanyId   int64  `json:"company_id"`   // Id cong ty
	Mst         string `json:"mst"`          // mã số thuế
	Identify    string `json:"identify"`     // CMT
	Note        string `json:"note"`         // Ghi chu them
}

func (item *CustomerInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item CustomerInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Booking Mush Pay (Must Pay)
type BookingMushPay struct {
	MushPay          int64 `json:"mush_pay"`
	TotalGolfFee     int64 `json:"total_golf_fee"`
	TotalServiceItem int64 `json:"total_service_item"`
}

func (item *BookingMushPay) UpdateAmount() {
	item.MushPay = item.TotalGolfFee + item.TotalServiceItem
}

func (item *BookingMushPay) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingMushPay) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Booking GolfFee
type BookingGolfFee struct {
	BookingUid string `json:"booking_uid"`
	PlayerName string `json:"player_name"`
	Bag        string `json:"bag"`
	CaddieFee  int64  `json:"caddie_fee"`
	BuggyFee   int64  `json:"buggy_fee"`
	GreenFee   int64  `json:"green_fee"`
	RoundIndex int    `json:"round_index"`
	PaidBy     string `json:"paid_by"`
}

type BookingTeeResponse struct {
	PartnerUid     string `json:"partner_uid"`
	CourseUid      string `json:"course_uid"`
	BookingDate    string `json:"booking_date"`
	TeeType        string `json:"tee_type"`
	TeePath        string `json:"tee_path"`
	TurnTime       string `json:"turn_time"`
	TeeTime        string `json:"tee_time"`
	TeeOffTime     string `json:"tee_off_time"`
	Bag            string `json:"bag"`
	Hole           int    `json:"hole"`
	GuestStyle     string `json:"guest_style"`
	GuestStyleName string `json:"guest_style_name"`

	// MemberCard
	CardId        string `json:"card_id"`
	MemberCardUid string `json:"member_card_uid"`

	// Thêm customer info
	CustomerBookingName  string                  `json:"customer_booking_name"`
	CustomerBookingPhone string                  `json:"customer_booking_phone"`
	CaddieId             int64                   `json:"caddie_id"`
	AgencyId             int64                   `json:"agency_id"`
	BookingCode          string                  `json:"booking_code"`
	BookingRestaurant    utils.BookingRestaurant `json:"booking_restaurant,omitempty"`
	BookingRetal         utils.BookingRental     `json:"booking_retal,omitempty"`
	Count                int64                   `json:"count"`
}

type ListBookingGolfFee []BookingGolfFee

func (item *ListBookingGolfFee) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingGolfFee) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Current Bag Price info
type BookingCurrentBagPriceDetail struct {
	Transfer    int64 `json:"transfer"`
	Debit       int64 `json:"debit"`
	GolfFee     int64 `json:"golf_fee"`
	Restaurant  int64 `json:"restaurant"`
	Kiosk       int64 `json:"kiosk"`
	Rental      int64 `json:"rental"`
	Proshop     int64 `json:"proshop"`
	Promotion   int64 `json:"promotion"`
	Amount      int64 `json:"amount"`
	AmountUsd   int64 `json:"amount_usd"`
	MainBagPaid int64 `json:"main_bag_paid"`
}

func (item *BookingCurrentBagPriceDetail) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingCurrentBagPriceDetail) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *BookingCurrentBagPriceDetail) UpdateAmount() {
	item.Amount = item.Transfer + item.Debit + item.GolfFee + item.Restaurant + item.Kiosk + item.Rental + item.Proshop + item.Promotion
}

// Booking Round
type BookingRound struct {
	Index         int    `json:"index"`
	CaddieFee     int64  `json:"caddie_fee"`
	BuggyFee      int64  `json:"buggy_fee"`
	GreenFee      int64  `json:"green_fee"`
	Hole          int    `json:"hole"`
	GuestStyle    string `json:"guest_style"` // Nếu là member Card thì lấy guest style của member Card, nếu không thì lấy guest style Của booking đó
	MemberCardId  string `json:"member_card_id"`
	MemberCardUid string `json:"member_card_uid"`
	Pax           int    `json:"pax"`
	TeeOffTime    int64  `json:"tee_off_time"`
}

type ListBookingRound []BookingRound

func (item *ListBookingRound) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingRound) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Agency info
type BookingAgency struct {
	Id             int64                 `json:"id"`
	Type           string                `json:"type"`
	AgencyId       string                `json:"agency_id"`       // Id Agency
	ShortName      string                `json:"short_name"`      // Ten ngắn Dai ly
	Category       string                `json:"category"`        // Category
	GuestStyle     string                `json:"guest_style"`     // Guest Style
	Name           string                `json:"name"`            // Ten Dai ly
	ContractDetail models.AgencyContract `json:"contract_detail"` // Thông tin đại lý
}

func (item *BookingAgency) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingAgency) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// type AgencyPaid struct {
// 	CaddieFee int64 `json:"caddie_fee"`
// 	BuggyFee  int64 `json:"buggy_fee"`
// 	GolfFee   int64 `json:"golf_fee"`
// 	Amount    int64 `json:"amount"`
// }

// func (item *AgencyPaid) Scan(v interface{}) error {
// 	return json.Unmarshal(v.([]byte), item)
// }

// func (item AgencyPaid) Value() (driver.Value, error) {
// 	return json.Marshal(&item)
// }

// Caddie Info
type BookingCaddie struct {
	Id       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Sex      bool   `json:"sex"`
	BirthDay int64  `json:"birth_day"`
	Group    string `json:"group"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

func (item *BookingCaddie) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingCaddie) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Buggy Info
type BookingBuggy struct {
	Id     int64  `json:"id"`
	Code   string `json:"code"`
	Number int    `json:"number"`
	Origin string `json:"origin"`
}

func (item *BookingBuggy) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingBuggy) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type NumberPeopleInFlight struct {
	FlightId int64 `json:"flight_id"`
	Total    int64 `json:"total"`
}

type BookingFeeOfBag struct {
	AgencyPaid        utils.ListBookingAgencyPayForBagData `json:"agency_paid,omitempty"`
	SubBags           utils.ListSubBag                     `json:"sub_bags,omitempty"`
	MushPayInfo       BookingMushPay                       `json:"mush_pay_info,omitempty"`
	ListServiceItems  []BookingServiceItemWithPaidInfo     `json:"list_service_items"`
	ListRoundOfSubBag []RoundOfBag                         `json:"list_round_of_sub_bag"`
	Rounds            []models.RoundPaidByMainBag          `json:"rounds"`
}

type AgencyCancelBookingList struct {
	BookingCode          string        `json:"booking_code"`
	AgencyId             int64         `json:"agency_id"`
	AgencyInfo           BookingAgency `json:"agency_info"`
	CustomerBookingName  string        `json:"customer_booking_name"`
	CustomerBookingPhone string        `json:"customer_booking_phone"`
	TeeOffTime           string        `json:"tee_off_time"`
	TeeTime              string        `json:"tee_time"`
	Hole                 int           `json:"hole"`
	NoteOfBag            string        `json:"note_of_bag"`
	NoteOfBooking        string        `json:"note_of_booking"`
	NumberPeople         int           `json:"number_people"`
	CancelBookingTime    int64         `json:"cancel_booking_time"` // Time cancel booking
}

type MainBagOfSubInfo struct {
	MainPaidRound1     bool
	MainPaidRound2     bool
	MainPaidRental     bool
	MainPaidProshop    bool
	MainPaidRestaurant bool
	MainPaidKiosk      bool
	MainPaidOtherFee   bool
	MainCheckOutTime   int64
	MainBagPaid        int64
}

// -------- Booking Logic --------

func (item *Booking) CheckDuplicatedCaddieInTeeTime(db *gorm.DB) bool {
	if item.TeeTime == "" {
		return false
	}

	booking := Booking{
		PartnerUid:  item.PartnerUid,
		CourseUid:   item.CourseUid,
		TeeTime:     item.TeeTime,
		BookingDate: item.BookingDate,
		CaddieId:    item.CaddieId,
	}

	errFind := booking.FindFirstNotCancel(db)
	return errFind == nil
}

// ----------- CRUD ------------
func (item *Booking) Create(db *gorm.DB, uid string) error {
	item.Model.Uid = uid
	now := time.Now()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *Booking) Update(db *gorm.DB) error {
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Booking) CreateBatch(db *gorm.DB, bookings []Booking) error {
	now := time.Now()
	for i := range bookings {
		c := &bookings[i]
		c.Model.CreatedAt = now.Unix()
		c.Model.UpdatedAt = now.Unix()
		c.Model.Status = constants.STATUS_ENABLE
	}

	return db.CreateInBatches(bookings, 100).Error
}

func (item *Booking) FindFirst(database *gorm.DB) error {
	db := database.Order("created_at desc")
	return db.Where(item).First(item).Error
}

func (item *Booking) FindFirstByUId(database *gorm.DB) (Booking, error) {
	errFSub := item.FindFirst(database)
	if errFSub == nil {
		if item.Bag != "" {
			booking := Booking{
				CourseUid:   item.CourseUid,
				PartnerUid:  item.PartnerUid,
				Bag:         item.Bag,
				BookingDate: item.BookingDate,
			}
			db := database.Order("created_at desc")
			db.Where(&booking).First(&booking)
			return booking, db.Error
		}
		return *item, errFSub
	}
	return Booking{}, errFSub
}

func (item *Booking) FindFirstWithJoin(database *gorm.DB) error {
	db := database.Order("created_at desc")
	return db.Where(item).First(item).Error
}

func (item *Booking) FindFirstNotCancel(db *gorm.DB) error {
	db = db.Where(item)
	db = db.Not("bag_status = ?", constants.BAG_STATUS_CANCEL)
	return db.Where(item).First(item).Error
}

func (item *Booking) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Booking{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Booking) FindAllBookingOTA(database *gorm.DB) ([]Booking, error) {
	db := database.Model(Booking{})
	list := []Booking{}

	db = db.Where("partner_uid = ?", item.PartnerUid)
	db = db.Where("booking_code = ?", item.BookingCode)
	db = db.Group("bill_code")

	db.Find(&list)
	return list, db.Error
}

func (item *Booking) FindAgencyCancelBooking(database *gorm.DB, page models.Page) ([]AgencyCancelBookingList, int64, error) {
	db := database.Model(Booking{})
	list := []AgencyCancelBookingList{}
	total := int64(0)

	db = db.Where("partner_uid = ?", item.PartnerUid)
	db = db.Where("course_uid = ?", item.CourseUid)

	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db = db.Group("booking_code")
	db = db.Where("agency_id <> ?", 0)
	db = db.Where("bag_status = ?", constants.BAG_STATUS_CANCEL)
	db = db.Select("bookings.*, COUNT(booking_code) as number_people")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Booking) FindAllBookingCheckIn(database *gorm.DB, bookingDate string) ([]Booking, error) {
	db := database.Model(Booking{})
	list := []Booking{}

	if bookingDate != "" {
		db = db.Where("booking_date = ?", bookingDate)
		db = db.Where("bag_status = ?", constants.BAG_STATUS_WAITING)
	}

	db.Find(&list)
	return list, db.Error
}

func (item *Booking) FindList(database *gorm.DB, page models.Page, from int64, to int64, agencyType string) ([]Booking, int64, error) {
	db := database.Model(Booking{})
	list := []Booking{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""
	// db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.AgencyId > 0 {
		db = db.Where("agency_id = ?", item.AgencyId)
	}

	if item.FlightId > 0 {
		db = db.Where("flight_id = ?", item.FlightId)
	}

	if item.BagStatus != "" {
		db = db.Where("bag_status = ?", item.BagStatus)
	}

	if item.CustomerName != "" {
		db = db.Where("customer_name LIKE ?", "%"+item.CustomerName+"%")
	}

	if item.Bag != "" {
		db = db.Where("bag LIKE ?", "%"+item.Bag+"%")
	}

	if agencyType != "" {
		db = db.Where("agency_info->'$.type' LIKE ?", "%"+agencyType+"%")
	}

	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}

	//Search With Time
	if from > 0 && to > 0 {
		db = db.Where("created_at between " + strconv.FormatInt(from, 10) + " and " + strconv.FormatInt(to, 10) + " ")
	}

	if from > 0 && to == 0 {
		db = db.Where("created_at > " + strconv.FormatInt(from, 10) + " ")
	}

	if from == 0 && to > 0 {
		db = db.Where("created_at < " + strconv.FormatInt(to, 10) + " ")
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
		db = db.Not("bag_status = ?", constants.BAG_STATUS_CANCEL)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Booking) FindBookingTeeTimeList(database *gorm.DB) ([]BookingTeeResponse, int64, error) {
	db := database.Model(Booking{})
	list := []BookingTeeResponse{}
	total := int64(0)
	item.Model.Status = ""
	db = db.Where(item)
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}
	if item.TeeTime != "" {
		db = db.Where("tee_time = ?", item.TeeTime)
	}
	db.Select("partner_uid,course_uid,booking_date,tee_type,tee_path,turn_time,tee_time,tee_off_time,hole,guest_style,guest_style_name,booking_code, COUNT(booking_code) as count").Group("booking_code")

	db.Count(&total)
	db.Find(&list)

	return list, total, db.Error
}

func (item *Booking) FindListForSubBag(database *gorm.DB) ([]BookingForSubBag, error) {
	db := database.Table("bookings")
	list := []BookingForSubBag{}
	status := item.Model.Status
	item.Model.Status = ""
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}

	bagStatus := []string{
		constants.BAG_STATUS_CHECK_OUT,
		constants.BAG_STATUS_BOOKING,
	}

	db = db.Where("bag_status NOT IN (?)", bagStatus)
	db.Find(&list)

	return list, db.Error
}

func (item *Booking) FindListWithBookingCode(database *gorm.DB) ([]Booking, error) {
	db := database.Table("bookings")
	list := []Booking{}
	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}
	db.Find(&list)
	return list, db.Error
}

/*
Find bookings in Flight
*/
func (item *Booking) FindListInFlight(database *gorm.DB) ([]Booking, error) {
	db := database.Table("bookings")
	list := []Booking{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.FlightId > 0 {
		db = db.Where("flight_id = ?", item.FlightId)
	}
	if item.Bag != "" {
		db = db.Where("bag = ?", item.Bag)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *Booking) Delete(db *gorm.DB) error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *Booking) FindForCaddieOnCourse(database *gorm.DB, InFlight string) []Booking {
	db := database.Model(Booking{})
	list := []Booking{}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BuggyId != 0 {
		db = db.Where("buggy_id = ?", item.BuggyId)
	}
	if item.CaddieId != 0 {
		db = db.Where("caddie_id = ?", item.CaddieId)
	}
	if item.Bag != "" {
		db = db.Where("bag COLLATE utf8mb4_general_ci LIKE ?", "%"+item.Bag+"%")
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}
	if item.CustomerName != "" {
		db = db.Where("customer_name COLLATE utf8mb4_general_ci LIKE LIKE ?", "%"+item.CustomerName+"%")
	}
	db = db.Where("bag_status = ?", constants.BAG_STATUS_WAITING)
	db = db.Not("caddie_status = ?", constants.BOOKING_CADDIE_STATUS_OUT)

	customerType := []string{
		constants.CUSTOMER_TYPE_NONE_GOLF,
		constants.CUSTOMER_TYPE_WALKING_FEE,
	}

	db = db.Where("customer_type NOT IN (?)", customerType)
	db = db.Order("created_at desc")

	if InFlight != "" {
		if InFlight == "0" {
			db = db.Not("flight_id > ?", 0)
		} else {
			db = db.Where("flight_id > ?", 0)
		}
	}
	db = db.Preload("CaddieBuggyInOut")
	db.Find(&list)
	return list
}

/*
Get List for Flight Data
*/
func (item *Booking) FindForFlightAll(database *gorm.DB, caddieCode string, caddieName string, numberPeopleInFlight *int64, page models.Page) []BookingForFlightRes {
	db := database.Table("bookings")
	list := []BookingForFlightRes{}
	listFlightWithNumberPeople := []int64{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}
	if caddieName != "" {
		db = db.Where("caddie_info->'$.name' LIKE ?", "%"+caddieName+"%")
	}

	if caddieCode != "" {
		db = db.Where("caddie_info->'$.code' = ?", caddieCode)
	}

	if item.CustomerName != "" {
		db = db.Where("customer_name = ?", item.CustomerName)
	}

	db = db.Where("flight_id > ?", 0)

	if numberPeopleInFlight != nil && *numberPeopleInFlight > 0 {
		listFlightR := []NumberPeopleInFlight{}
		db2 := datasources.GetDatabase().Table("bookings")
		db2.Select("COUNT(flight_id) as total,flight_id").Group("flight_id").Having("COUNT(flight_id) = ?", *numberPeopleInFlight)
		db2.Find(&listFlightR)
		for _, item := range listFlightR {
			listFlightWithNumberPeople = append(listFlightWithNumberPeople, item.FlightId)
		}
		db.Where("flight_id in (?) ", listFlightWithNumberPeople)
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	err := db.Error
	if err != nil {
		log.Println("Booking FindForFlightAll err ", err.Error())
	}
	return list
}

/*
For report MainBag SubBag
*/
func (item *Booking) FindListForReportForMainBagSubBag(database *gorm.DB) ([]BookingForReportMainBagSubBags, error) {
	db := database.Table("bookings")
	list := []BookingForReportMainBagSubBags{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	db = db.Where("booking_date = ?", item.BookingDate)

	// db.Where("bag <> ''")
	// db.Where("moved_flight = 0 AND added_round = 0")
	db.Group("bag")
	db.Order("created_at desc")

	db.Find(&list)

	return list, db.Error
}

/*
For report List Service Items
*/
func (item *Booking) FindListServiceItems(database *gorm.DB, param GetListBookingWithListServiceItems, page models.Page) ([]BookingForListServiceIems, int64, error) {
	db := database.Table("bookings")
	list := []Booking{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("bookings.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("bookings.course_uid = ?", item.CourseUid)
	}

	if param.GolfBag != "" {
		db = db.Where("bookings.bag = ?", param.GolfBag)
	}

	if param.PlayerName != "" {
		db = db.Where("bookings.customer_name LIKE ?", "%"+param.PlayerName+"%")
	}

	if param.ServiceType != "" {
		db = db.Where("booking_service_items.type = ?", param.ServiceType)
	}
	db = db.Joins("RIGHT JOIN booking_service_items ON booking_service_items.booking_uid = bookings.uid")
	db = db.Order("booking_service_items.created_at desc")
	db = db.Group("booking_service_items.bill_code")
	db.Count(&total)
	db = db.Preload("ListServiceItems")

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	listItems := []BookingForListServiceIems{}
	for _, data := range list {
		item := BookingForListServiceIems{
			PartnerUid:       data.PartnerUid,
			CourseUid:        data.CourseUid,
			BookingDate:      data.BookingDate,
			Bag:              data.Bag,
			ListServiceItems: data.ListServiceItems,
			CheckInTime:      data.CheckInTime,
			CustomerName:     data.CustomerName,
		}
		listItems = append(listItems, item)
	}

	return listItems, total, db.Error
}

func (item *Booking) ResetCaddieBuggy() {
	item.CaddieId = 0
	item.CaddieInfo = BookingCaddie{}
	item.CaddieStatus = ""

	item.BuggyId = 0
	item.BuggyInfo = BookingBuggy{}
}

/*
Lấy ra tee time index còn avaible
*/
func (item *Booking) FindTeeTimeIndexAvaible(database *gorm.DB) utils.ListInt {
	db := database.Table("bookings")
	list := []Booking{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}
	if item.TeeTime != "" {
		db = db.Where("tee_time = ?", item.TeeTime)
	}
	if item.TeeType != "" {
		db = db.Where("tee_type = ?", item.TeeType)
	}

	db.Find(&list)

	listIndex := utils.ListInt{}

	isAdd0 := true
	isAdd1 := true
	isAdd2 := true
	isAdd3 := true

	for _, v := range list {
		if v.RowIndex != nil {
			rIndex := v.RowIndex
			if *rIndex == 0 {
				isAdd0 = false
			} else if *rIndex == 1 {
				isAdd1 = false
			} else if *rIndex == 2 {
				isAdd2 = false
			} else if *rIndex == 3 {
				isAdd3 = false
			}
		}
	}

	if isAdd0 {
		listIndex = append(listIndex, 0)
	}
	if isAdd1 {
		listIndex = append(listIndex, 1)
	}
	if isAdd2 {
		listIndex = append(listIndex, 2)
	}
	if isAdd3 {
		listIndex = append(listIndex, 3)
	}

	return listIndex
}

/*
Find MainBag of Bag
*/
func (item *Booking) FindMainBag(database *gorm.DB) ([]Booking, error) {
	db := database.Table("bookings")
	list := []Booking{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}
	if item.Bag != "" {
		db = db.Where("JSON_SEARCH(sub_bags ->'$[*]', 'one', ?) IS NOT NULL", item.Bag)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *Booking) FindTopMember(database *gorm.DB, memberType, dateType, date string) ([]map[string]interface{}, error) {
	db := database.Table("bookings")
	list := []map[string]interface{}{}

	if memberType == constants.TOP_MEMBER_TYPE_PLAY {
		db.Select("card_id, customer_name, COUNT(*) as play_count")
	} else {
		db.Select("card_id, customer_name, SUM(mush_pay_info->'$.mush_pay') as sales")
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if dateType == constants.TOP_MEMBER_DATE_TYPE_MONTH {
		db = db.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%Y-%m') = ?", date)
	} else if dateType == constants.TOP_MEMBER_DATE_TYPE_WEEK {
		db = db.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%u') = ?", date)
	}
	// else if dateType == constants.TOP_MEMBER_DATE_TYPE_DAY {
	// 	db = db.Where("booking_date = ?", date)
	// }

	db = db.Where("customer_type = ?", constants.BOOKING_CUSTOMER_TYPE_MEMBER)

	db = db.Where("check_in_time > 0")

	db = db.Where("check_out_time > 0")

	db.Group("card_id")

	if memberType == constants.TOP_MEMBER_TYPE_PLAY {
		db.Order("play_count desc")
	} else {
		db.Order("sales desc")
	}

	db.Limit(10)
	db.Find(&list)

	return list, db.Error
}

func (item *Booking) ReportBookingRevenue(database *gorm.DB, bookingType, date string) ([]map[string]interface{}, error) {
	db := database.Table("bookings")
	list := []map[string]interface{}{}

	db.Select("SUM(mush_pay_info->'$.mush_pay') as revenue")

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db = db.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%Y-%m') = ?", date)

	db = db.Where("check_in_time > 0")

	db = db.Where("check_out_time > 0")

	if bookingType == constants.BOOKING_CUSTOMER_TYPE_AGENCY {
		db = db.Where("(customer_type = ? OR customer_type = ?)", constants.BOOKING_CUSTOMER_TYPE_OTA, constants.BOOKING_CUSTOMER_TYPE_TRADITIONAL)
	} else if bookingType == constants.BOOKING_CUSTOMER_TYPE_GUEST {
		db = db.Where("customer_type = ?", constants.BOOKING_CUSTOMER_TYPE_GUEST)
	} else if bookingType == constants.BOOKING_CUSTOMER_TYPE_MEMBER {
		db = db.Where("customer_type = ?", constants.BOOKING_CUSTOMER_TYPE_MEMBER)
	} else if bookingType == constants.BOOKING_CUSTOMER_TYPE_VISITOR {
		db = db.Where("customer_type = ?", constants.BOOKING_CUSTOMER_TYPE_VISITOR)
	} else {
		customerTypes := []string{
			constants.BOOKING_CUSTOMER_TYPE_OTA,
			constants.BOOKING_CUSTOMER_TYPE_TRADITIONAL,
			constants.BOOKING_CUSTOMER_TYPE_MEMBER,
			constants.BOOKING_CUSTOMER_TYPE_GUEST,
			constants.BOOKING_CUSTOMER_TYPE_VISITOR,
		}

		db = db.Where("customer_type NOT IN (?) ", customerTypes)
	}

	db.Group("course_uid")

	db.Find(&list)

	return list, db.Error
}
