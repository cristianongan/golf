package request

import "encoding/json"

type CreateBookingWaitingBody struct {
	BookingDate string `json:"booking_date"`                   // dd/mm/yyyy
	CmsUser     string `json:"cms_user"`                       // Acc Operator Tạo (Bỏ lấy theo token)
	PartnerUid  string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid   string `json:"course_uid" binding:"required"`  // San Golf
	CourseType  string `json:"course_type"`

	Hole     int    `json:"hole"`      // Số hố check
	TeeType  string `json:"tee_type"`  // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeePath  string `json:"tee_path"`  // MORNING, NOON, NIGHT (k required cái này vì có case checking k qua booking)
	TurnTime string `json:"turn_time"` // Ex: 16:26 (k required cái này vì có case checking k qua booking)
	TeeTime  string `json:"tee_time"`  // Ex: 16:26 Tee time là thời gian tee off dự kiến (k required cái này vì có case checking k qua booking)

	// Guest booking
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name

	// Member Card
	MemberCardUid        *string `json:"member_card_uid"`
	CustomerName         string  `json:"customer_name"`          // Tên khách hàng
	CustomerBookingName  string  `json:"customer_booking_name"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string  `json:"customer_booking_phone"` // SDT khách hàng đặt booking
	CustomerUid          string  `json:"customer_uid"`

	Note string `json:"note"` // Note of Booking

	//Agency
	AgencyId    int64   `json:"agency_id"`
	CaddieCode  *string `json:"caddie_code"`
	BookingCode string  `json:"booking_code"`

	MemberUidOfGuest  *string `json:"member_uid_of_guest"`
	MemberNameOfGuest string  `json:"member_name_of_guest"`
	Id                int64   `json:"id"`
}

type GetListBookingWaitingForm struct {
	PageRequest
	PartnerUid    string `form:"partner_uid"`
	CourseUid     string `form:"course_uid"`
	BookingDate   string `form:"booking_date"`
	PlayerName    string `form:"player_name"`
	BookingCode   string `form:"booking_code"`
	PlayerContact string `form:"player_contact"`
}

type UpdateBookingWaiting struct {
	AccessoriesId *int    `json:"accessories_id"`
	Amount        *int    `json:"amount"`
	Note          *string `json:"note"`
}

type DeleteBookingWaiting struct {
	PartnerUid  string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid   string `json:"course_uid" binding:"required"`  // San Golf
	BookingCode string `json:"booking_code" binding:"required"`
	TeeType     string `json:"tee_type" binding:"required"`
	TeeTime     string `json:"tee_time" binding:"required"`
	CourseType  string `json:"course_type" binding:"required"`
	BookingDate string `json:"booking_date" binding:"required"`
}

type CreateBatchBookingWaitingBody struct {
	BookingList ListCreateBatchBookingWaitingBody `json:"booking_list"`
}

type ListCreateBatchBookingWaitingBody []CreateBookingWaitingBody

func (item *ListCreateBatchBookingWaitingBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}
