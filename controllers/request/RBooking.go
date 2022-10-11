package request

import (
	"database/sql/driver"
	"encoding/json"
	model_booking "start/models/booking"
	"start/utils"
)

type GetListBookingSettingGroupForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
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
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	Bag         string `form:"bag"`
	From        int64  `form:"from"`
	To          int64  `form:"to"`
	BookingDate string `form:"booking_date"`
	BookingCode string `form:"booking_code"`
	AgencyId    int64  `form:"agency_id"`
	BagStatus   string `form:"bag_status"`
	PlayerName  string `form:"player_name"`
	FlightId    int64  `form:"flight_id"`
	AgencyType  string `form:"agency_type"`
	TeeTime     string `form:"tee_time"`
}

type GetListBookingWithSelectForm struct {
	PageRequest
	PartnerUid     string  `form:"partner_uid"`
	CourseUid      string  `form:"course_uid"`
	BookingDate    string  `form:"booking_date"`
	BookingCode    string  `form:"booking_code"`
	InitType       string  `form:"init_type"`
	AgencyId       int64   `form:"agency_id"`
	IsAgency       string  `form:"is_agency"`
	Status         string  `form:"status"`
	FromDate       string  `form:"from_date"`
	ToDate         string  `form:"to_date"`
	GolfBag        string  `form:"golf_bag"`
	IsToday        string  `form:"is_today"`
	BookingUid     string  `form:"booking_uid"`
	IsFlight       string  `form:"is_flight"`
	BagStatus      string  `form:"bag_status"`
	HaveBag        *string `form:"have_bag"`
	CaddieCode     string  `form:"caddie_code"`
	HasBookCaddie  string  `form:"has_book_caddie"`
	PlayerName     string  `form:"player_name"`
	HasFlightInfo  string  `form:"has_flight_info"`
	HasCaddieInOut string  `form:"has_caddie_in_out"`
	FlightId       int64   `form:"flight_id"`
	TeeType        string  `form:"tee_type"`
	IsCheckIn      string  `form:"is_check_in"`
	GuestStyleName string  `form:"guest_style_name"`
	PlayerOrBag    string  `form:"player_or_bag"`
}

type GetListBookingWithListServiceItems struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Type       string `form:"type"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
	GolfBag    string `form:"golf_bag"`
	PlayerName string `form:"player_name"`
}

type GetListBookingTeeTimeForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date"`
	TeeTime     string `form:"tee_time"`
}

type CancelAllBookingBody struct {
	PartnerUid  string `form:"partner_uid" json:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" json:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date" json:"booking_date"`
	BookingCode string `form:"booking_code" json:"booking_code"`
	TeeTime     string `form:"tee_time" json:"tee_time"`
	Reason      string `form:"reason" json:"reason"`
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
	Bag         string `json:"bag"`       // Golf Bag
	Hole        int    `json:"hole"`      // Số hố
	TeeType     string `json:"tee_type"`  // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeePath     string `json:"tee_path"`  // MORNING, NOON, NIGHT (k required cái này vì có case checking k qua booking)
	TurnTime    string `json:"turn_time"` // Ex: 16:26 (k required cái này vì có case checking k qua booking)
	TeeTime     string `json:"tee_time"`  // Ex: 16:26 Tee time là thời gian tee off dự kiến (k required cái này vì có case checking k qua booking)
	RowIndex    *int   `json:"row_index"` // index trong Flight

	// Guest booking
	GuestStyle           string `json:"guest_style"`            // Guest Style
	GuestStyleName       string `json:"guest_style_name"`       // Guest Style Name
	CustomerName         string `json:"customer_name"`          // Tên khách hàng
	CustomerBookingName  string `json:"customer_booking_name"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string `json:"customer_booking_phone"` // SDT khách hàng đặt booking
	CustomerIdentify     string `json:"customer_identify"`      // passport/cccd

	NoteOfBooking string `json:"note_of_booking"` // Note of Booking

	// Member Card
	MemberCardUid string `json:"member_card_uid"`
	IsCheckIn     bool   `json:"is_check_in"`

	MemberUidOfGuest string `json:"member_uid_of_guest"` // Member của Guest đến chơi cùng

	//Agency
	AgencyId           int64                   `json:"agency_id"`
	CustomerUid        string                  `json:"customer_uid"`
	CaddieCode         string                  `json:"caddie_code"`
	BookingRestaurant  utils.BookingRestaurant `json:"booking_restaurant"`
	BookingRetal       utils.BookingRental     `json:"booking_retal"`
	BookingCode        string                  `json:"booking_code"`
	BookingCodePartner string                  `json:"booking_code_partner"`
	BookingSourceId    string                  `json:"booking_source_id"`
	BookingOtaId       int64                   `json:"booking_ota_id"`
	BookMark           bool
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

	MemberCardUid    string `json:"member_card_uid"`     // Member Card
	AgencyId         int64  `json:"agency_id"`           // Agency id
	MemberUidOfGuest string `json:"member_uid_of_guest"` // Member của Guest đến chơi cùng
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
	Hole        int      `json:"hole"`
}

type UpdateBooking struct {
	model_booking.Booking
	CaddieCode string `json:"caddie_code"`
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
