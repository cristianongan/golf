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
}

type GetListBookingWithSelectForm struct {
	PageRequest
	PartnerUid  string  `form:"partner_uid"`
	CourseUid   string  `form:"course_uid"`
	Bag         string  `form:"bag"`
	BookingDate string  `form:"booking_date"`
	BookingCode string  `form:"booking_code"`
	InitType    string  `form:"init_type"`
	AgencyId    int64   `form:"agency_id"`
	IsAgency    string  `form:"is_agency"`
	Status      string  `form:"status"`
	FromDate    string  `form:"from_date"`
	ToDate      string  `form:"to_date"`
	GolfBag     string  `form:"golf_bag"`
	IsToday     string  `form:"is_today"`
	BookingUid  string  `form:"booking_uid"`
	IsFlight    string  `form:"is_flight"`
	BagStatus   string  `form:"bag_status"`
	HaveBag     *string `form:"have_bag"`
}

type GetListBookingTeeTimeForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date"`
	TeeTime     string `form:"tee_time"`
}

// Tạo Tee booking
// Guest Booking
// Member Booking
type CreateBookingBody struct {
	BookingDate string `json:"booking_date"`                   // dd/mm/yyyy
	CmsUser     string `json:"cms_user"`                       // Acc Operator Tạo
	PartnerUid  string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid   string `json:"course_uid" binding:"required"`  // San Golf
	Bag         string `json:"bag"`                            // Golf Bag
	Hole        int    `json:"hole"`                           // Số hố
	TeeType     string `json:"tee_type" binding:"required"`    // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath     string `json:"tee_path" binding:"required"`    // MORNING, NOON, NIGHT
	TurnTime    string `json:"turn_time" binding:"required"`   // Ex: 16:26
	TeeTime     string `json:"tee_time" binding:"required"`    // Ex: 16:26 Tee time là thời gian tee off dự kiến
	RowIndex    int    `json:"row_index"`                      // index trong Flight

	// Guest booking
	GuestStyle           string `json:"guest_style"`            // Guest Style
	GuestStyleName       string `json:"guest_style_name"`       // Guest Style Name
	CustomerName         string `json:"customer_name"`          // Tên khách hàng
	CustomerBookingName  string `json:"customer_booking_name"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string `json:"customer_booking_phone"` // SDT khách hàng đặt booking

	// Member Card
	MemberCardUid string `json:"member_card_uid"`
	IsCheckIn     bool   `json:"is_check_in"`

	//Agency
	AgencyId          int64                   `json:"agency_id"`
	CustomerUid       string                  `json:"customer_uid"`
	CaddieCode        string                  `json:"caddie_code"`
	BookingRestaurant utils.BookingRestaurant `json:"booking_restaurant"`
	BookingRetal      utils.BookingRental     `json:"booking_retal"`
	BookingCode       string                  `form:"booking_code"`
	BookingSourceId   string                  `json:"booking_source_id"`
}

type CreateBatchBookingBody struct {
	BookingList ListCreateBookingBody `json:"booking_list"`
	IsWaiting   bool                  `json:"is_waiting"` // booking ở trạng thái chờ
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
	Bag    string `json:"bag" binding:"required"` // Golf Bag
	Locker string `json:"locker"`
	Hole   int    `json:"hole"` // Số hố

	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name
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
	TeeTime     string   `json:"tee_time" validate:"required"`
	Hole        int      `json:"hole"`
}

type UpdateBooking struct {
	model_booking.Booking
	CaddieCode string `json:"caddie_code"`
}
