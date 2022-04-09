package request

import "start/utils"

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
}

type GetListBookingForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Bag        string `form:"bag"`
}

// Tạo Tee booking
// Guest Booking
// Member Booking
type CreateBookingBody struct {
	CmsUser    string `json:"cms_user"`                       // Acc Operator Tạo
	PartnerUid string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid  string `json:"course_uid" binding:"required"`  // San Golf
	Bag        string `json:"bag"`                            // Golf Bag
	Hole       int    `json:"hole"`                           // Số hố
	TeeType    string `json:"tee_type"`                       // 1, 1A, 1B, 1C, 10, 10A, 10B
	TeePath    string `json:"tee_path"`                       // MORNING, NOON, NIGHT
	TurnTime   string `json:"turn_time"`                      // Ex: 16:26
	TeeTime    string `json:"tee_time"`                       // Ex: 16:26 Tee time là thời gian tee off dự kiến
	RowIndex   int    `json:"row_index"`                      // index trong Flight

	// Guest booking
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name
	CustomerName   string `json:"customer_name"`    // Tên khách hàng

	// Member Card
	MemberCardUid string `json:"member_card_uid"`
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

type CheckInBody struct {
	BookingBaseBody
	Bag    string `json:"bag" binding:"required"` // Golf Bag
	Locker string `json:"locker"`
	Hole   int    `json:"hole"` // Số hố
}
