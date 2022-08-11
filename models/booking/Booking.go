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
)

// Booking
// omitempty: xứ lý khi các field trả về rỗng
type Booking struct {
	models.Model
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf

	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022

	Bag            string `json:"bag" gorm:"type:varchar(100);index"`         // Golf Bag
	Hole           int    `json:"hole"`                                       // Số hố
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

	CurrentBagPrice  BookingCurrentBagPriceDetail `json:"current_bag_price,omitempty" gorm:"type:json"`  // Thông tin phí++: Tính toán lại phí Service items, Tiền cho Subbag
	ListGolfFee      ListBookingGolfFee           `json:"list_golf_fee,omitempty" gorm:"type:json"`      // Thông tin List Golf Fee, Main Bag, Sub Bag
	ListServiceItems ListBookingServiceItems      `json:"list_service_items,omitempty" gorm:"type:json"` // List service item: rental, proshop, restaurant, kiosk
	MushPayInfo      BookingMushPay               `json:"mush_pay_info,omitempty" gorm:"type:json"`      // Mush Pay info
	Rounds           ListBookingRound             `json:"rounds,omitempty" gorm:"type:json"`             // List Rounds: Sẽ sinh golf Fee với List GolfFee
	OtherPaids       utils.ListOtherPaid          `json:"other_paids,omitempty" gorm:"type:json"`        // Other Paids

	// Note          string `json:"note" gorm:"type:varchar(500)"`            // Note
	NoteOfBag     string `json:"note_of_bag" gorm:"type:varchar(500)"`     // Note of Bag
	NoteOfBooking string `json:"note_of_booking" gorm:"type:varchar(500)"` // Note of Booking
	LockerNo      string `json:"locker_no" gorm:"type:varchar(100)"`       // Locker mã số tủ gửi đồ
	ReportNo      string `json:"report_no" gorm:"type:varchar(200)"`       // Report No
	CancelNote    string `json:"cancel_note" gorm:"type:varchar(300)"`     // Cancel note

	CmsUser    string `json:"cms_user" gorm:"type:varchar(100)"`     // Cms User
	CmsUserLog string `json:"cms_user_log" gorm:"type:varchar(200)"` // Cms User Log

	// TODO
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
	AgencyId   int64         `json:"agency_id" gorm:"index"`
	AgencyInfo BookingAgency `json:"agency_info" gorm:"type:json"`

	// Subs bags
	SubBags utils.ListSubBag `json:"sub_bags,omitempty" gorm:"type:json"` // List Sub Bags

	// Main bags
	MainBags utils.ListSubBag `json:"main_bags,omitempty" gorm:"type:json"` // List Main Bags, thêm main bag sẽ thanh toán những cái gì
	// Main bug for Pay: Mặc định thanh toán all, Nếu có trong list này thì k thanh toán
	MainBagNoPay utils.ListString `json:"main_bag_no_pay,omitempty" gorm:"type:json"` // Main Bag không thanh toán những phần này
	SubBagNote   string           `json:"sub_bag_note" gorm:"type:varchar(500)"`      // Note of SubBag

	InitType string `json:"init_type" gorm:"type:varchar(50);index"` // BOOKING: Tạo booking xong checkin, CHECKIN: Check In xong tạo Booking luôn

	CaddieInOut       []CaddieInOutNote       `json:"caddie_in_out" gorm:"foreignKey:BookingUid;references:Uid"`
	BookingCode       string                  `json:"booking_code" gorm:"type:varchar(100);index"` // cho case tạo nhiều booking có cùng booking code
	BookingRestaurant utils.BookingRestaurant `json:"booking_restaurant,omitempty" gorm:"type:json"`
	BookingRetal      utils.BookingRental     `json:"booking_retal,omitempty" gorm:"type:json"`
	BookingSourceId   string                  `json:"booking_source_id" gorm:"type:varchar(50);index"`

	MemberUidOfGuest  string `json:"member_uid_of_guest" gorm:"type:varchar(50);index"` // Member của Guest đến chơi cùng
	MemberNameOfGuest string `json:"member_name_of_guest" gorm:"type:varchar(200)"`     // Member của Guest đến chơi cùng

	HasBookCaddie bool  `json:"has_book_caddie" gorm:"default:0"`
	TimeOutFlight int64 `json:"time_out_flight,omitempty"`
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

type CaddieInOutNote CaddieInOutNoteForBooking

type CaddieInOutNoteForBooking struct {
	models.ModelId
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	BookingUid string `json:"booking_uid"`
	CaddieId   int64  `json:"caddie_id"`
	CaddieCode string `json:"caddie_code"`
	Note       string `json:"note"`
	Type       string `json:"type"`
	Hole       int    `json:"hole"`
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
func (item *Booking) UpdateBookingMainBag() error {
	if item.MainBags == nil || len(item.MainBags) == 0 {
		return errors.New("invalid main bags")
	}
	mainBagBookingUid := item.MainBags[0].BookingUid
	mainBagBooking := Booking{}
	mainBagBooking.Uid = mainBagBookingUid
	errFindMainB := mainBagBooking.FindFirst()
	if errFindMainB != nil {
		return errFindMainB
	}

	if mainBagBooking.ListGolfFee == nil {
		mainBagBooking.ListGolfFee = ListBookingGolfFee{}
	}

	// Update lại cho Main Bag Booking
	// Check GolfFee
	if item.ListGolfFee != nil {
		idxTemp := -1
		for i, gf := range mainBagBooking.ListGolfFee {
			if gf.BookingUid == item.Uid {
				idxTemp = i
			}
		}
		if idxTemp == -1 {
			// Chưa có thì thêm vào
			mainBagBooking.ListGolfFee = append(mainBagBooking.ListGolfFee, item.GetCurrentBagGolfFee())
		} else {
			// Update cái mới
			mainBagBooking.ListGolfFee[idxTemp] = item.GetCurrentBagGolfFee()
		}
	}

	// Udp list service items
	if mainBagBooking.ListServiceItems == nil {
		mainBagBooking.ListServiceItems = ListBookingServiceItems{}
	}

	if item.ListServiceItems != nil && len(item.ListServiceItems) > 0 {
		for _, v := range item.ListServiceItems {
			// Check cùng booking và cùng item id
			idxTemp := -1
			if len(mainBagBooking.ListServiceItems) > 0 {
				for i, v1 := range mainBagBooking.ListServiceItems {
					if v1.BookingUid == v.BookingUid && v1.ItemId == v.ItemId {
						idxTemp = i
					}
				}
			}

			if idxTemp == -1 {
				// Chưa có thì thêm vào List
				mainBagBooking.ListServiceItems = append(mainBagBooking.ListServiceItems, v)
			} else {
				// Update cái mới
				mainBagBooking.ListServiceItems[idxTemp] = v
			}
		}
	}

	// Udp lại mush Pay
	mainBagBooking.UpdateMushPay()

	errUdp := mainBagBooking.Update()
	if errUdp != nil {
		return errUdp
	}

	return nil
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

func (item *Booking) AddRound(memberCardUid string, golfFee models.GolfFee) error {
	lengthRound := len(item.Rounds)

	if memberCardUid == "" {
		// Guest

	}

	// Member
	memberCard := models.MemberCard{}
	memberCard.Uid = memberCardUid
	errFind := memberCard.FindFirst()
	if errFind != nil {
		return errFind
	}

	bookingRound := BookingRound{
		Index:         lengthRound + 1,
		Hole:          item.Hole,
		Pax:           1,
		MemberCardId:  memberCard.CardId,
		MemberCardUid: memberCardUid,
		TeeOffTime:    time.Now().Unix(),
	}
	bookingRound.CaddieFee = utils.GetFeeFromListFee(golfFee.CaddieFee, bookingRound.Hole)
	bookingRound.GreenFee = utils.GetFeeFromListFee(golfFee.GreenFee, bookingRound.Hole)
	bookingRound.BuggyFee = utils.GetFeeFromListFee(golfFee.BuggyFee, bookingRound.Hole)

	item.Rounds = append(item.Rounds, bookingRound)

	return nil
}

func (item *Booking) UpdateBagGolfFee() {
	if len(item.ListGolfFee) > 0 {
		item.ListGolfFee[0].Bag = item.Bag
	}
}

// Udp MushPay
func (item *Booking) UpdateMushPay() {
	mushPay := BookingMushPay{}

	totalGolfFee := int64(0)
	for _, v := range item.ListGolfFee {
		totalGolfFee += (v.BuggyFee + v.CaddieFee + v.GreenFee)
	}
	mushPay.TotalGolfFee = totalGolfFee

	// SubBag

	// Sub Service Item của current Bag
	for _, v := range item.ListServiceItems {
		isNeedPay := true
		if len(item.MainBagNoPay) > 0 {
			for _, v1 := range item.MainBagNoPay {
				// TODO: Tính Fee cho sub bag fee
				if v1 == constants.MAIN_BAG_FOR_PAY_SUB_NEXT_ROUNDS {
				} else if v1 == constants.MAIN_BAG_FOR_PAY_SUB_FIRST_ROUND {
				} else if v1 == constants.MAIN_BAG_FOR_PAY_SUB_OTHER_FEE {
				} else {
					if v1 == v.Type {
						isNeedPay = false
					}
				}
			}
		}
		if isNeedPay {
			mushPay.TotalServiceItem += v.Amount
		}
	}

	mushPay.MushPay = mushPay.TotalGolfFee + mushPay.TotalServiceItem
	item.MushPayInfo = mushPay
}

// Udp lại giá cho Booking
// Udp sub bag price
func (item *Booking) UpdatePriceDetailCurrentBag() {
	priceDetail := BookingCurrentBagPriceDetail{}

	if len(item.ListGolfFee) > 0 {
		priceDetail.GolfFee = item.ListGolfFee[0].BuggyFee + item.ListGolfFee[0].CaddieFee + item.ListGolfFee[0].GreenFee
	}

	for _, serviceItem := range item.ListServiceItems {
		if serviceItem.BookingUid == item.Uid {
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
func (item *Booking) IsDuplicated(checkTeeTime, checkBag bool) (bool, error) {
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
		}

		errFind := booking.FindFirstNotCancel()
		if errFind == nil || booking.Uid != "" {
			return true, errors.New("Duplicated TeeTime")
		}
	}

	//Check Bag đã tồn tại
	if checkBag {
		if item.Bag != "" {
			booking := Booking{
				PartnerUid:  item.PartnerUid,
				CourseUid:   item.CourseUid,
				BookingDate: item.BookingDate,
				Bag:         item.Bag,
			}
			errBagFind := booking.FindFirst()
			if errBagFind == nil || booking.Uid != "" {
				return true, errors.New("Duplicated Bag")
			}
		}
	}

	return false, nil
}

// ----------- CRUD ------------
func (item *Booking) Create(uid string) error {
	item.Model.Uid = uid
	now := time.Now()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Booking) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Booking) CreateBatch(bookings []Booking) error {
	now := time.Now()
	for i := range bookings {
		c := &bookings[i]
		c.Model.CreatedAt = now.Unix()
		c.Model.UpdatedAt = now.Unix()
		c.Model.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.CreateInBatches(bookings, 100).Error
}

func (item *Booking) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Booking) FindFirstNotCancel() error {
	db := datasources.GetDatabase()
	db = db.Where(item)
	db = db.Not("bag_status = ?", constants.BAG_STATUS_CANCEL)
	return db.Where(item).First(item).Error
}

func (item *Booking) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Booking) FindAllBookingCheckIn(bookingDate string) ([]Booking, error) {
	db := datasources.GetDatabase().Model(Booking{})
	list := []Booking{}

	if bookingDate != "" {
		db = db.Where("booking_date = ?", bookingDate)
		db = db.Where("bag_status = ?", constants.BAG_STATUS_IN)
	}

	db.Find(&list)
	return list, db.Error
}

func (item *Booking) FindList(page models.Page, from int64, to int64, agencyType string) ([]Booking, int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
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

func (item *Booking) FindBookingTeeTimeList() ([]BookingTeeResponse, int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
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

func (item *Booking) FindListForSubBag() ([]BookingForSubBag, error) {
	db := datasources.GetDatabase().Table("bookings")
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
	if item.BagStatus != "" {
		db = db.Where("bag_status = ?", item.BagStatus)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *Booking) FindListWithBookingCode() ([]Booking, error) {
	db := datasources.GetDatabase().Table("bookings")
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
func (item *Booking) FindListInFlight() ([]Booking, error) {
	db := datasources.GetDatabase().Table("bookings")
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

func (item *Booking) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *Booking) FindForCaddieOnCourse(InFlight string) []Booking {
	db := datasources.GetDatabase().Model(Booking{})
	list := []Booking{}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	//if item.BookingDate == "" {
	//	dateDisplay, errDate := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	//	if errDate == nil {
	//		item.BookingDate = dateDisplay
	//	} else {
	//		log.Println("FindForCaddieOnCourse BookingDate err ", errDate.Error())
	//	}
	//}
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
	db = db.Where("bag_status = ?", constants.BAG_STATUS_IN)
	db = db.Not("caddie_status = ?", constants.BOOKING_CADDIE_STATUS_OUT)
	if InFlight != "" {
		if InFlight == "0" {
			db = db.Not("flight_id > ?", 0)
		} else {
			db = db.Where("flight_id > ?", 0)
		}
	}

	db.Find(&list)
	return list
}

/*
Get List for Flight Data
*/
func (item *Booking) FindForFlightAll(caddieCode string, caddieName string, numberPeopleInFlight *int64, page models.Page) []BookingForFlightRes {
	db := datasources.GetDatabase().Table("bookings")
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
func (item *Booking) FindListForReportForMainBagSubBag() ([]BookingForReportMainBagSubBags, error) {
	db := datasources.GetDatabase().Table("bookings")
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
