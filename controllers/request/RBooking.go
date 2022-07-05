package request

import (
	"database/sql/driver"
	"encoding/json"
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
	TeeType     string `json:"tee_type"`                       // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath     string `json:"tee_path"`                       // MORNING, NOON, NIGHT
	TurnTime    string `json:"turn_time"`                      // Ex: 16:26
	TeeTime     string `json:"tee_time"`                       // Ex: 16:26 Tee time là thời gian tee off dự kiến
	RowIndex    int    `json:"row_index"`                      // index trong Flight

	// Guest booking
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name
	CustomerName   string `json:"customer_name"`    // Tên khách hàng

	// Member Card
	MemberCardUid string `json:"member_card_uid"`
	IsCheckIn     bool   `json:"is_check_in"`

	//Agency
	AgencyId    int64  `json:"agency_id"`
	CustomerUid string `json:"customer_uid"`
}

type BookingBaseBody struct {
	BookingUid string `json:"booking_uid"`
	CmsUser    string `json:"cms_user"`
	Note       string `json:"note"`
}

// Thêm service item vào booking
type AddServiceItemToBooking struct {
	BookingBaseBody
	ServiceItems utils.ListBookingServiceItems `json:"service_items"`
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

type AddRoundBody struct {
	BookingBaseBody
	MemberCardId string `json:"member_card_id"`
	GuestStyle   string `json:"guest_style"`
}

// ------ Other Paid --------
type AddOtherPaidBody struct {
	BookingBaseBody
	OtherPaids utils.ListOtherPaid `json:"other_paids"`
}

type CancelBookingBody struct {
	BookingBaseBody
}

type MovingBookingBody struct {
	BookUidList *[]string `json:"booking_uid_list"`
	BookingDate *string   `json:"booking_date"`
	TeeType     *string   `json:"tee_type"`
	TeeTime     *string   `json:"tee_time"`
	Hole        *int      `json:"hole"`
}
