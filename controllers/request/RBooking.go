package request

import (
	"database/sql/driver"
	"encoding/json"
	model_booking "start/models/booking"
	"start/utils"
)

type GetListBookingSettingGroupForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
}

type GetListBookingSettingForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	GroupId    int64  `form:"group_id"`
	OnDate     string `form:"on_date"`
}

type GetListBookingForm struct {
	PageRequest
	PartnerUid       string `form:"partner_uid" binding:"required"`
	CourseUid        string `form:"course_uid" binding:"required"`
	Bag              string `form:"bag"`
	From             int64  `form:"from"`
	To               int64  `form:"to"`
	BookingDate      string `form:"booking_date"`
	BookingCode      string `form:"booking_code"`
	AgencyId         int64  `form:"agency_id"`
	BagStatus        string `form:"bag_status"`
	PlayerName       string `form:"player_name"`
	FlightId         int64  `form:"flight_id"`
	AgencyType       string `form:"agency_type"`
	TeeTime          string `form:"tee_time"`
	HasRoundOfSubBag string `form:"has_round_of_sub_bag"`
}

type GetListBookingWithSelectForm struct {
	PageRequest
	PartnerUid      string  `form:"partner_uid" binding:"required"`
	CourseUid       string  `form:"course_uid" binding:"required"`
	BookingDate     string  `form:"booking_date"`
	TeeTime         string  `form:"tee_time"`
	BookingCode     string  `form:"booking_code"`
	InitType        string  `form:"init_type"`
	AgencyId        int64   `form:"agency_id"`
	AgencyName      string  `form:"agency_name"`
	IsAgency        string  `form:"is_agency"`
	Status          string  `form:"status"`
	FromDate        string  `form:"from_date"`
	ToDate          string  `form:"to_date"`
	GolfBag         string  `form:"golf_bag"`
	IsToday         string  `form:"is_today"`
	BookingUid      string  `form:"booking_uid"`
	IsFlight        string  `form:"is_flight"`
	BagStatus       string  `form:"bag_status"`
	HaveBag         *string `form:"have_bag"`
	CaddieCode      string  `form:"caddie_code"`
	CaddieName      string  `form:"caddie_name"`
	HasBookCaddie   string  `form:"has_book_caddie"`
	PlayerName      string  `form:"player_name"`
	HasFlightInfo   string  `form:"has_flight_info"`
	HasCaddieInOut  string  `form:"has_caddie_in_out"`
	FlightId        int64   `form:"flight_id"`
	TeeType         string  `form:"tee_type"`
	CourseType      string  `form:"course_type"`
	IsCheckIn       string  `form:"is_check_in"`
	GuestStyleName  string  `form:"guest_style_name"`
	PlayerOrBag     string  `form:"player_or_bag"`
	IsGroupBillCode bool    `form:"is_group_bill_code"`
	HasBuggy        string  `form:"has_buggy"`
	HasCaddie       string  `form:"has_caddie"`
	CustomerUid     string  `form:"customer_uid"`
	CustomerType    string  `form:"customer_type"`
	BuggyCode       string  `form:"buggy_code"`
	GuestStyle      string  `form:"guest_style"`
	GuestType       string  `form:"guest_type"`
	CheckInCode     string  `form:"check_in_code"`
}

type GetListBookingWithListServiceItems struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Type       string `form:"type"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
	GolfBag    string `form:"golf_bag"`
	PlayerName string `form:"player_name"`
}

type GetListBookingTeeTimeForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date"`
	TeeTime     string `form:"tee_time"`
}

type CancelAllBookingBody struct {
	PartnerUid  string `form:"partner_uid" json:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" json:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date" json:"booking_date" binding:"required"`
	BookingCode string `form:"booking_code" json:"booking_code"`
	TeeTime     string `form:"tee_time" json:"tee_time"`
	TeeType     string `form:"tee_type" json:"tee_type"`
	CourseType  string `form:"course_type" json:"course_type"`
	Reason      string `form:"reason" json:"reason"`
}

type FinishBookingBody struct {
	PartnerUid  string `form:"partner_uid" json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid   string `form:"course_uid" json:"course_uid" binding:"required"`   // San Golf
	Bag         string `form:"bag" json:"bag"`
	BillNo      string `form:"bill_no" json:"bill_no"`
	BookingDate string `form:"booking_date" json:"booking_date"`
}

// Tạo Tee booking
// Guest Booking
// Member Booking
type CreateBookingBody struct {
	BookingDate string `json:"booking_date"`                   // dd/mm/yyyy
	CmsUser     string `json:"cms_user"`                       // Acc Operator Tạo (Bỏ lấy theo token)
	PartnerUid  string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid   string `json:"course_uid" binding:"required"`  // San Golf
	CourseType  string `json:"course_type"`
	Bag         string `json:"bag"`          // Golf Bag
	HoleBooking int    `json:"hole_booking"` // Số hố
	Hole        int    `json:"hole"`         // Số hố check
	TeeType     string `json:"tee_type"`     // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeePath     string `json:"tee_path"`     // MORNING, NOON, NIGHT (k required cái này vì có case checking k qua booking)
	TurnTime    string `json:"turn_time"`    // Ex: 16:26 (k required cái này vì có case checking k qua booking)
	TeeTime     string `json:"tee_time"`     // Ex: 16:26 Tee time là thời gian tee off dự kiến (k required cái này vì có case checking k qua booking)
	RowIndex    *int   `json:"row_index"`    // index trong Flight

	// Guest booking
	GuestStyle           string  `json:"guest_style"`            // Guest Style
	GuestStyleName       string  `json:"guest_style_name"`       // Guest Style Name
	CustomerName         string  `json:"customer_name"`          // Tên khách hàng
	CustomerBookingEmail *string `json:"customer_booking_email"` // Email khách hàng
	CustomerBookingName  string  `json:"customer_booking_name"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string  `json:"customer_booking_phone"` // SDT khách hàng đặt booking
	CustomerIdentify     string  `json:"customer_identify"`      // passport/cccd
	Nationality          string  `json:"nationality"`            // Nationality

	NoteOfBooking string `json:"note_of_booking"` // Note of Booking

	// Member Card
	MemberCardUid string `json:"member_card_uid"`
	IsCheckIn     bool   `json:"is_check_in"`

	MemberUidOfGuest string `json:"member_uid_of_guest"` // Member của Guest đến chơi cùng

	//Agency
	AgencyId           int64                   `json:"agency_id"`
	CustomerUid        string                  `json:"customer_uid"`
	CaddieCode         *string                 `json:"caddie_code"`
	CaddieCheckIn      *string                 `json:"caddie_checkin"`
	BookingRestaurant  utils.BookingRestaurant `json:"booking_restaurant"`
	BookingRetal       utils.BookingRental     `json:"booking_retal"`
	BookingCode        string                  `json:"booking_code"`
	BookingCodePartner string                  `json:"booking_code_partner"`
	BookingSourceId    string                  `json:"booking_source_id"`
	BookingOtaId       int64                   `json:"booking_ota_id"`
	LockerNo           string                  `json:"locker_no"` // Locker mã số tủ gửi đồ
	ReportNo           string                  `json:"report_no"` // Report No
	IsPrivateBuggy     *bool                   `json:"is_private_buggy"`
	FeeInfo            *AgencyFeeInfo          `json:"fee_info"`
	AgencyPaidAll      *bool                   `json:"agency_paid_all"`
	BookingWaitingId   int64                   `json:"booking_waiting_id"`
	BookMark           bool
	BookFromOTA        bool
	BookingTeeTime     bool
}

/*
Update để đồng bộ với cách lưu trong redis và database:
database mysql đang chia tee_type: 1, 10m, course_type: A,B,C
redis đang lưu teeType: 1A, 1B, 1C,...
*/
func (item *CreateBookingBody) UpdateTeeType(teeType string) {
	if teeType == "1A" {
		item.TeeType = "1"
		item.CourseType = "A"
	} else if teeType == "1B" {
		item.TeeType = "1"
		item.CourseType = "B"
	} else if teeType == "1C" {
		item.TeeType = "1"
		item.CourseType = "C"
	} else if teeType == "10A" {
		item.TeeType = "10"
		item.CourseType = "A"
	} else if teeType == "10B" {
		item.TeeType = "10"
		item.CourseType = "B"
	} else if teeType == "10c" {
		item.TeeType = "10"
		item.CourseType = "C"
	} else {
		item.TeeType = teeType
		item.CourseType = "A"
	}
}

type UpdateAgencyOrMemberCardToBooking struct {
	PartnerUid    string
	CourseUid     string
	AgencyId      int64
	BUid          string
	Bag           string
	CustomerName  string
	Hole          int
	MemberCardUid string
}

type GolfFeeGuestyleParam struct {
	Uid          string
	Rate         string
	Bag          string
	CustomerName string
	CaddieFee    int64
	BuggyFee     int64
	GreenFee     int64
	Hole         int
}

type CreateBatchBookingBody struct {
	BookingList ListCreateBookingBody `json:"booking_list"`
}

type ListCreateBookingBody []CreateBookingBody

func (item *ListCreateBookingBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListCreateBookingBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type AgencyFeeInfo struct {
	GolfFee       int64 `json:"golf_fee"`
	BuggyFee      int64 `json:"buggy_fee"`
	CaddieFee     int64 `json:"caddie_fee"`
	OddCarFee     int64 `json:"odd_car_fee"`
	PrivateCarFee int64 `json:"private_car_fee"`
}
type BookingBaseBody struct {
	BookingUid string `json:"booking_uid"`
	CmsUser    string `json:"cms_user"`
	Note       string `json:"note"`
}

// Thêm service item vào booking
type AddServiceItemToBooking struct {
	BookingBaseBody
	ServiceItems model_booking.ListBookingServiceItems `json:"service_items"`
}

// GO: Ghép flight

// Thêm Subbag
type AddSubBagToBooking struct {
	BookingBaseBody
	SubBags utils.ListSubBag `json:"sub_bags"`
}

// Edit Subbag
type EditSubBagToBooking struct {
	BookingBaseBody
	SubBags ListEditSubBagBooking `json:"sub_bags"`
}
type ListEditSubBagBooking []EditSubBagBooking

func (item *ListEditSubBagBooking) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListEditSubBagBooking) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type EditSubBagBooking struct {
	BookingUid string `json:"booking_uid"`
	PlayerName string `json:"player_name"`
	SubBagNote string `json:"sub_bag_note"` // Note of SubBag
	IsOut      bool   `json:"is_out"`
}

type CheckInBody struct {
	BookingBaseBody
	Bag            string `json:"bag" binding:"required"` // Golf Bag
	Locker         string `json:"locker"`
	Hole           int    `json:"hole"`             // Số hố
	CourseType     string `json:"course_type"`      // Sân nào : A,B,C
	TeeType        string `json:"tee_type"`         // Tee nào : 1, 10
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name
	CustomerName   string `json:"customer_name"`    // Player Name

	MemberCardUid    *string        `json:"member_card_uid"`
	AgencyId         int64          `json:"agency_id"`           // Agency id
	MemberUidOfGuest string         `json:"member_uid_of_guest"` // Member của Guest đến chơi cùng
	FeeInfo          *AgencyFeeInfo `json:"fee_info"`            // Golf Fee cho case agency
	AgencyPaidAll    *bool          `json:"agency_paid_all"`

	CaddieCode string `json:"caddie_code"`
}

//type AddRoundBody struct {
//	BookingBaseBody
//	MemberCardId string `json:"member_card_id"`
//	GuestStyle   string `json:"guest_style"`
//}

// ------ Other Paid --------
type AddOtherPaidBody struct {
	BookingBaseBody
	OtherPaids utils.ListOtherPaid `json:"other_paids"`
}

type CancelBookingBody struct {
	BookingBaseBody
}

type MovingBookingBody struct {
	BookUidList []string `json:"booking_uid_list" validate:"required"`
	BookingDate string   `json:"booking_date" validate:"required"`
	TeeType     string   `json:"tee_type" validate:"required"`
	CourseType  string   `json:"course_type"`
	TeeTime     string   `json:"tee_time" validate:"required"`
	TeePath     string   `json:"tee_path" validate:"required"`
	TurnTime    string   `json:"turn_time" validate:"required"`
	Hole        int      `json:"hole"`
}

type UpdateBooking struct {
	BookingDate string  `json:"booking_date"`                   // dd/mm/yyyy
	CmsUser     string  `json:"cms_user"`                       // Acc Operator Tạo (Bỏ lấy theo token)
	PartnerUid  string  `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid   string  `json:"course_uid" binding:"required"`  // San Golf
	CourseType  string  `json:"course_type"`
	Bag         *string `json:"bag"`          // Golf Bag
	Hole        int     `json:"hole"`         // Số hố
	HoleBooking int     `json:"hole_booking"` // Số hố khi booking
	TeeType     string  `json:"tee_type"`     // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeePath     string  `json:"tee_path"`     // MORNING, NOON, NIGHT (k required cái này vì có case checking k qua booking)
	TurnTime    string  `json:"turn_time"`    // Ex: 16:26 (k required cái này vì có case checking k qua booking)
	TeeTime     string  `json:"tee_time"`     // Ex: 16:26 Tee time là thời gian tee off dự kiến (k required cái này vì có case checking k qua booking)
	RowIndex    *int    `json:"row_index"`    // index trong Flight

	// Guest booking
	GuestStyle           string  `json:"guest_style"`            // Guest Style
	GuestStyleName       string  `json:"guest_style_name"`       // Guest Style Name
	CustomerName         string  `json:"customer_name"`          // Tên khách hàng
	CustomerBookingEmail *string `json:"customer_booking_email"` // Email khách hàng
	CustomerBookingName  string  `json:"customer_booking_name"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string  `json:"customer_booking_phone"` // SDT khách hàng đặt booking
	CustomerIdentify     string  `json:"customer_identify"`      // passport/cccd
	Nationality          string  `json:"nationality"`            // Nationality

	NoteOfBooking *string `json:"note_of_booking"` // Note of Booking

	// Member Card
	MemberCardUid *string `json:"member_card_uid"`
	IsCheckIn     bool    `json:"is_check_in"`

	MemberUidOfGuest string `json:"member_uid_of_guest"` // Member của Guest đến chơi cùng

	//Agency
	AgencyId           int64                   `json:"agency_id"`
	CustomerUid        string                  `json:"customer_uid"`
	CaddieCheckIn      *string                 `json:"caddie_checkin"`
	CaddieCode         *string                 `json:"caddie_code"`
	BookingRestaurant  utils.BookingRestaurant `json:"booking_restaurant"`
	BookingRetal       utils.BookingRental     `json:"booking_retal"`
	BookingCode        string                  `json:"booking_code"`
	BookingCodePartner string                  `json:"booking_code_partner"`
	BookingSourceId    string                  `json:"booking_source_id"`
	BookingOtaId       int64                   `json:"booking_ota_id"`
	LockerNo           *string                 `json:"locker_no"` // Locker mã số tủ gửi đồ
	ReportNo           *string                 `json:"report_no"` // Report No
	IsPrivateBuggy     *bool                   `json:"is_private_buggy"`
	FeeInfo            *AgencyFeeInfo          `json:"fee_info"`
	AgencyPaidAll      *bool                   `json:"agency_paid_all"`
	NoteOfBag          *string                 `json:"note_of_bag" gorm:"type:varchar(500)"`    // Note of Bag
	NoteOfGo           *string                 `json:"note_of_go" gorm:"type:varchar(500)"`     // Note khi trong GO
	MainBagPay         utils.ListString        `json:"main_bag_pay,omitempty" gorm:"type:json"` // Main Bag không thanh toán những phần này ở sub bag này
}

type ChangeBookingHole struct {
	Hole           int    `json:"hole" validate:"required"`
	TypeChangeHole string `json:"type_stop"`
	NoteOfBag      string `json:"note_of_bag" validate:"required"`
}

type UpdateAgencyFeeBookingCommon struct {
	Uid          string
	CourseUid    string
	AgencyId     int64
	Bag          string
	CheckInTime  int64
	CustomerName string
	Hole         int
}

type LockBill struct {
	PartnerUid string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid  string `json:"course_uid" binding:"required"`  // San Golf
	Bag        string `json:"bag" binding:"required"`
	LockBill   *bool  `json:"lock_bill" binding:"required"`
}

type UndoCheckOut struct {
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid" binding:"required"`
	Bag         string `json:"bag" binding:"required"`
	BookingDate string `json:"booking_date" binding:"required"`
}

type ReportPaymentBagStatus struct {
	PartnerUid    string `form:"partner_uid" binding:"required"`
	CourseUid     string `form:"course_uid" binding:"required"`
	BookingDate   string `form:"booking_date"`
	PaymentStatus string `form:"payment_status"`
}

type ReportBookingPlayers struct {
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid" binding:"required"`
	BookingDate string `json:"booking_date"`
}

type GetCaddieBookingCancel struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date" binding:"required"`
	CaddieCode  string `form:"caddie_code"`
	CaddieName  string `form:"caddie_name"`
}

type SendInforGuestBody struct {
	SendMethod  string        `json:"send_method" binding:"required"`
	ListBooking []ItemBooking `json:"list_booking" binding:"required"`
}

type ItemBooking struct {
	Uid                  string `json:"uid"`
	CustomerName         string `json:"customer_name"`
	CustomerBookingPhone string `json:"customer_booking_phone"`
	CustomerBookingEmail string `json:"customer_booking_email"`
	CaddieBooking        string `json:"caddie_booking"`
}
