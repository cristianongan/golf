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

	BagStatus    string `json:"bag_status" gorm:"type:varchar(50);index"` // Bag status
	CheckInTime  int64  `json:"check_in_time"`                            // Time Check In
	CheckOutTime int64  `json:"check_out_time"`                           // Time Check Out
	TeeType      string `json:"tee_type" gorm:"type:varchar(50);index"`   // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath      string `json:"tee_path" gorm:"type:varchar(50);index"`   // MORNING, NOON, NIGHT
	TurnTime     string `json:"turn_time" gorm:"type:varchar(30)"`        // Ex: 16:26
	TeeTime      string `json:"tee_time" gorm:"type:varchar(30)"`         // Ex: 16:26 Tee time là thời gian tee off dự kiến
	TeeOffTime   string `json:"tee_off_time" gorm:"type:varchar(30)"`     // Ex: 16:26 Là thời gian thực tế phát bóng
	RowIndex     *int   `json:"row_index"`                                // index trong Flight

	CurrentBagPrice BookingCurrentBagPriceDetail `json:"current_bag_price,omitempty" gorm:"type:json"` // Thông tin phí++: Tính toán lại phí Service items, Tiền cho Subbag
	ListGolfFee     ListBookingGolfFee           `json:"list_golf_fee,omitempty" gorm:"type:json"`     // Thông tin List Golf Fee, Main Bag, Sub Bag
	MushPayInfo     BookingMushPay               `json:"mush_pay_info,omitempty" gorm:"type:json"`     // Mush Pay info
	OtherPaids      utils.ListOtherPaid          `json:"other_paids,omitempty" gorm:"type:json"`       // Other Paids

	// Note          string `json:"note" gorm:"type:varchar(500)"`            // Note
	NoteOfBag     string `json:"note_of_bag" gorm:"type:varchar(500)"`     // Note of Bag
	NoteOfBooking string `json:"note_of_booking" gorm:"type:varchar(500)"` // Note of Booking
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

	CaddieInOut       []CaddieBuggyInOut      `json:"caddie_in_out" gorm:"foreignKey:BookingUid;references:Uid"`
	BuggyInOut        []BuggyInOut            `json:"buggy_in_out" gorm:"foreignKey:BookingUid;references:Uid"`
	BookingCode       string                  `json:"booking_code" gorm:"type:varchar(100);index"` // cho case tạo nhiều booking có cùng booking code
	BookingRestaurant utils.BookingRestaurant `json:"booking_restaurant,omitempty" gorm:"type:json"`
	BookingRetal      utils.BookingRental     `json:"booking_retal,omitempty" gorm:"type:json"`
	BookingSourceId   string                  `json:"booking_source_id" gorm:"type:varchar(50);index"`

	MemberUidOfGuest  string `json:"member_uid_of_guest" gorm:"type:varchar(50);index"` // Member của Guest đến chơi cùng
	MemberNameOfGuest string `json:"member_name_of_guest" gorm:"type:varchar(200)"`     // Member của Guest đến chơi cùng

	HasBookCaddie bool   `json:"has_book_caddie" gorm:"default:0"`
	TimeOutFlight int64  `json:"time_out_flight,omitempty"`                // Thời gian out Flight
	BillCode      string `json:"bill_code" gorm:"type:varchar(100);index"` // hỗ trợ query tính giá
	SeparatePrice bool   `json:"separate_price" gorm:"default:0"`          // Giá riêng

	ListServiceItems []BookingServiceItem `json:"list_service_items,omitempty" gorm:"-:migration"` // List service item: rental, proshop, restaurant, kiosk
	ShowCaddieBuggy  *bool                `json:"show_caddie_buggy" gorm:"default:1"`              // Sau add round thì không hiển thị caddie buggy
	IsPrivateBuggy   *bool                `json:"is_private_buggy" gorm:"default:0"`               // Bag có dùng buggy riêng không
	// Rounds           ListBookingRound             `json:"rounds,omitempty" gorm:"type:json"`             // List Rounds: Sẽ sinh golf Fee với List GolfFee
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

type BookingForListServiceIems struct {
	PartnerUid       string               `json:"partner_uid"`                                                              // Hang Golf
	CourseUid        string               `json:"course_uid"`                                                               // San Golf
	BookingDate      string               `json:"booking_date"`                                                             // Ex: 06/11/2022
	Bag              string               `json:"bag"`                                                                      // Golf Bag
	ListServiceItems []BookingServiceItem `json:"list_service_items,omitempty" gorm:"foreignKey:BookingUid;references:Uid"` // List service item: rental, proshop, restaurant, kiosk
	CheckInTime      int64                `json:"check_in_time"`
	CustomerName     string               `json:"customer_name"`
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
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	BookingUid string `json:"booking_uid"`
	CaddieId   int64  `json:"caddie_id"`
	CaddieCode string `json:"caddie_code"`
	BuggyId    int64  `json:"buggy_id"`
	BuggyCode  string `json:"buggy_code"`
	Note       string `json:"note"`
	CaddieType string `json:"caddie_type"`
	BuggyType  string `json:"buggy_type"`
	Hole       int    `json:"hole"`
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
	Transfer   int64 `json:"transfer"`
	Debit      int64 `json:"debit"`
	GolfFee    int64 `json:"golf_fee"`
	Restaurant int64 `json:"restaurant"`
	Kiosk      int64 `json:"kiosk"`
	Rental     int64 `json:"rental"`
	Proshop    int64 `json:"proshop"`
	Promotion  int64 `json:"promotion"`
	Amount     int64 `json:"amount"`
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

// -------- Booking Logic --------
/*
	Lấy service item của main bag và sub bag nếu có
*/
func (item *Booking) FindServiceItems(db *gorm.DB) {
	//MainBag
	listServiceItems := ListBookingServiceItems{}
	serviceGolfs := BookingServiceItem{
		BillCode: item.BillCode,
	}
	listGolfService, _ := serviceGolfs.FindAll(db)
	if len(listGolfService) > 0 {
		for _, v := range listGolfService {
			// Check trạng thái bill
			serviceCart := models.ServiceCart{}
			serviceCart.Id = v.ServiceBill

			errSC := serviceCart.FindFirst(db)
			if errSC != nil {
				log.Println("FindFristServiceCart errSC", errSC.Error())
				return
			}

			if serviceCart.BillStatus != constants.POS_BILL_STATUS_OUT &&
				serviceCart.BillStatus != constants.POS_BILL_STATUS_ACTIVE &&
				serviceCart.BillStatus != constants.RES_BILL_STATUS_BOOKING &&
				serviceCart.BillStatus != constants.RES_BILL_STATUS_OUT &&
				serviceCart.BillStatus != constants.RES_BILL_STATUS_CANCEL {
				listServiceItems = append(listServiceItems, v)
			}
		}
	}

	//Check Subbag
	listTemp := ListBookingServiceItems{}
	if item.SubBags != nil && len(item.SubBags) > 0 {
		for _, v := range item.SubBags {
			serviceGolfsTemp := BookingServiceItem{
				BillCode: v.BillCode,
			}
			listGolfServiceTemp, _ := serviceGolfsTemp.FindAll(db)

			for _, v1 := range listGolfServiceTemp {
				isCanAdd := false
				if item.MainBagPay != nil && len(item.MainBagPay) > 0 {
					for _, v2 := range item.MainBagPay {
						// Check trạng thái bill
						serviceCart := models.ServiceCart{}
						serviceCart.Id = v1.ServiceBill

						errSC := serviceCart.FindFirst(db)
						if errSC != nil {
							log.Println("FindFristServiceCart errSC", errSC.Error())
							return
						}

						// Check trong MainBag có trả mới add
						if v2 == v1.Type && serviceCart.BillStatus != constants.POS_BILL_STATUS_OUT {
							isCanAdd = true
						}
					}
				}

				if isCanAdd {
					listTemp = append(listTemp, v1)
				}
			}
		}
	}

	listServiceItems = append(listServiceItems, listTemp...)

	item.ListServiceItems = listServiceItems
}

func (item *Booking) GetCurrentBagGolfFee() BookingGolfFee {
	golfFee := BookingGolfFee{}
	if item.ListGolfFee == nil {
		return golfFee
	}
	if len(item.ListGolfFee) <= 0 {
		return golfFee
	}

	return item.ListGolfFee[0]
}

func (item *Booking) GetTotalServicesFee() int64 {
	total := int64(0)
	if item.ListServiceItems != nil {
		for _, v := range item.ListServiceItems {
			temp := v.Amount
			total += temp
		}
	}

	return total
}

func (item *Booking) GetTotalGolfFee() int64 {
	total := int64(0)
	if item.ListGolfFee != nil {
		for _, v := range item.ListGolfFee {
			golfFeeTemp := v.BuggyFee + v.CaddieFee + v.GreenFee
			total += golfFeeTemp
		}
	}

	return total
}

func (item *Booking) UpdateBagGolfFee() {
	if len(item.ListGolfFee) > 0 {
		item.ListGolfFee[0].Bag = item.Bag
	}
}

// Udp MushPay
func (item *Booking) UpdateMushPay(db *gorm.DB) {
	mushPay := BookingMushPay{}

	totalGolfFee := int64(0)
	for _, v := range item.ListGolfFee {
		totalGolfFee += (v.BuggyFee + v.CaddieFee + v.GreenFee)
	}
	mushPay.TotalGolfFee = totalGolfFee

	// SubBag

	// Sub Service Item của current Bag
	// Get item for current Bag
	// update lại lấy service items mới
	item.FindServiceItems(db)
	for _, v := range item.ListServiceItems {
		isNeedPay := false
		if len(item.MainBagPay) > 0 && item.BillCode != v.BillCode {
			for _, v1 := range item.MainBagPay {
				// TODO: Tính Fee cho sub bag fee
				if v1 == constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS {
					// Next Round
				} else if v1 == constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND {
					// First Round
				} else if v1 == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE {
					// Other Fee cũng nằm trong service items
					if v1 == v.Type {
						isNeedPay = true
					}
				} else {
					if v1 == v.Type {
						isNeedPay = true
					}
				}
			}
		} else {
			if item.BillCode == v.BillCode {
				isNeedPay = true
			}
		}
		if isNeedPay {
			mushPay.TotalServiceItem += v.Amount
		}
	}

	mushPay.MushPay = mushPay.TotalGolfFee + mushPay.TotalServiceItem
	item.MushPayInfo = mushPay
}

/*
Update mush price bag have main bag
*/
func (item *Booking) UpdatePriceForBagHaveMainBags(db *gorm.DB) {
	mainBook := Booking{}
	mainBook.Uid = item.MainBags[0].BookingUid
	errFMB := mainBook.FindFirst(db)
	if errFMB != nil {
		log.Println("UpdatePriceForBagHaveMainBags errFMB", errFMB.Error())
		return
	}
	if item.MainBags == nil || len(item.MainBags) == 0 {
		return
	}
	listPay := mainBook.MainBagPay
	mushPay := BookingMushPay{}

	totalGolfFee := int64(0)

	// Check xem main bag có trả golf fee cho sub bag không
	// Check thanh toán first round
	isConFR := utils.ContainString(listPay, constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND)
	// Check thanh toán next round
	isConNR := utils.ContainString(listPay, constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS)
	for i, v := range item.ListGolfFee {
		if i == 0 {
			if isConFR < 0 {
				// Nếu main k thanh toán FR cho sub thì add vào sub
				totalGolfFee += (v.BuggyFee + v.CaddieFee + v.GreenFee)
			}
		} else {
			if isConNR < 0 {
				// Nếu main k thanh toán NR cho sub thì add vào sub
				totalGolfFee += (v.BuggyFee + v.CaddieFee + v.GreenFee)
			}
		}
	}

	// Tính total golf fee cho sub
	mushPay.TotalGolfFee = totalGolfFee

	item.FindServiceItems(db)
	for _, v := range item.ListServiceItems {
		isCon := utils.ContainString(listPay, v.Type)
		if isCon < 0 {
			// Main bag không thanh toán cho sub bag thì cộng vào
			mushPay.TotalServiceItem += v.Amount
		}
	}

	// Mush pay
	mushPay.MushPay = mushPay.TotalGolfFee + mushPay.TotalServiceItem
	item.MushPayInfo = mushPay

	//Udp current Bag price
	priceDetail := BookingCurrentBagPriceDetail{}

	if isConFR >= 0 {
		//TODO: Có thể phải check cho case nhiều round
		priceDetail.GolfFee = 0
	}
	for _, serviceItem := range item.ListServiceItems {
		isCon := utils.ContainString(listPay, serviceItem.Type)
		if isCon < 0 {
			// Main bag không thanh toán cho sub bag thì cộng vào
			if serviceItem.BillCode == item.BillCode {
				// Udp service detail cho booking uid
				if serviceItem.Type == constants.GOLF_SERVICE_RENTAL {
					priceDetail.Rental += serviceItem.Amount
				}
				if serviceItem.Type == constants.GOLF_SERVICE_PROSHOP {
					priceDetail.Proshop += serviceItem.Amount
				}
				if serviceItem.Type == constants.GOLF_SERVICE_RESTAURANT {
					priceDetail.Restaurant += serviceItem.Amount
				}
				if serviceItem.Type == constants.GOLF_SERVICE_KIOSK {
					priceDetail.Kiosk += serviceItem.Amount
				}
			}
		}
	}

	priceDetail.UpdateAmount()

	item.CurrentBagPrice = priceDetail

	//Udp price for main bag
	// trả cho thằng con
	listGolfFeeTemp := mainBook.ListGolfFee
	isIndex := -1
	for i, v := range mainBook.ListGolfFee {
		if v.BookingUid == item.Uid {
			isIndex = i
		}
	}
	if isIndex == -1 {
		//Chua dc add
		if isConFR >= 0 {
			mainBook.ListGolfFee = append(listGolfFeeTemp, item.ListGolfFee[0])
		}
	} else {
		if isConFR >= 0 {
			// add them vao
			mainBook.ListGolfFee[isIndex] = item.ListGolfFee[0]
		} else {
			// remove di
			listTempGF1 := ListBookingGolfFee{}
			for _, v := range mainBook.ListGolfFee {
				if v.BookingUid != item.Uid {
					listTempGF1 = append(listTempGF1, v)
				}
			}
			mainBook.ListGolfFee = listTempGF1
		}
	}
	mainBook.UpdateMushPay(db)
	mainBook.UpdatePriceDetailCurrentBag(db)
	errUdpMB := mainBook.Update(db)
	if errUdpMB != nil {
		log.Println("UpdatePriceForBagHaveMainBags errUdpMB", errUdpMB.Error())
	}
}

// Udp lại giá cho Booking
// Udp sub bag price
func (item *Booking) UpdatePriceDetailCurrentBag(db *gorm.DB) {
	priceDetail := BookingCurrentBagPriceDetail{}

	if len(item.ListGolfFee) > 0 {
		priceDetail.GolfFee = item.ListGolfFee[0].BuggyFee + item.ListGolfFee[0].CaddieFee + item.ListGolfFee[0].GreenFee
	}
	item.FindServiceItems(db)
	for _, serviceItem := range item.ListServiceItems {
		if serviceItem.BillCode == item.BillCode {
			// Udp service detail cho booking uid
			if serviceItem.Type == constants.GOLF_SERVICE_RENTAL {
				priceDetail.Rental += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_PROSHOP {
				priceDetail.Proshop += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_RESTAURANT {
				priceDetail.Restaurant += serviceItem.Amount
			}
			if serviceItem.Type == constants.GOLF_SERVICE_KIOSK {
				priceDetail.Kiosk += serviceItem.Amount
			}
		}
	}

	priceDetail.UpdateAmount()

	item.CurrentBagPrice = priceDetail
}

// Check Duplicated
func (item *Booking) IsDuplicated(db *gorm.DB, checkTeeTime, checkBag bool) (bool, error) {
	//Check Bag đã tồn tại trước
	if checkBag {
		if item.Bag != "" {
			booking := Booking{
				PartnerUid:  item.PartnerUid,
				CourseUid:   item.CourseUid,
				BookingDate: item.BookingDate,
				Bag:         item.Bag,
			}
			errBagFind := booking.FindFirst(db)
			if errBagFind == nil || booking.Uid != "" {
				return true, errors.New("Duplicated Bag")
			}
		}
	}

	if item.TeeTime == "" {
		return false, nil
	}
	//Check turn time đã tồn tại
	if checkTeeTime {
		booking := Booking{
			PartnerUid:  item.PartnerUid,
			CourseUid:   item.CourseUid,
			TeeTime:     item.TeeTime,
			TurnTime:    item.TurnTime,
			BookingDate: item.BookingDate,
			RowIndex:    item.RowIndex,
			TeeType:     item.TeeType,
			CourseType:  item.CourseType,
		}

		errFind := booking.FindFirstNotCancel(db)
		if errFind == nil || booking.Uid != "" {
			return true, errors.New("Duplicated TeeTime")
		}
	}

	return false, nil
}

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
		constants.BAG_STATUS_TIMEOUT,
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
		db = db.Where("bag = ?", item.Bag)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}
	if item.CustomerName != "" {
		db = db.Where("customer_name LIKE ?", "%"+item.CustomerName+"%")
	}
	db = db.Where("bag_status = ?", constants.BAG_STATUS_WAITING)
	db = db.Not("caddie_status = ?", constants.BOOKING_CADDIE_STATUS_OUT)

	customerType := []string{
		constants.CUSTOMER_TYPE_NONE_GOLF,
		constants.CUSTOMER_TYPE_WALKING_FEE,
	}

	db = db.Where("customer_type NOT IN (?)", customerType)

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
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
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
	db = db.Joins("INNER JOIN booking_service_items ON booking_service_items.booking_uid = bookings.uid").Group("bookings.bag")

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
