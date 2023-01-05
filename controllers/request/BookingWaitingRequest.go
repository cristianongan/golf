package request

import "start/utils"

type CreateBookingWaiting struct {
	PartnerUid    string           `json:"partner_uid"`
	CourseUid     string           `json:"course_uid"`
	BookingTime   string           `json:"booking_time"`
	TeeTime       string           `json:"tee_time"`
	PlayerName    string           `json:"player_name"`
	PlayerContact string           `json:"player_contact"`
	PeopleList    utils.ListString `json:"people_list"`
	Note          string           `json:"note"`
}

type GetListBookingWaitingForm struct {
	PageRequest
	PartnerUid    string `form:"partner_uid"`
	CourseUid     string `form:"course_uid"`
	Date          string `form:"date"`
	PlayerName    string `form:"player_name"`
	BookingCode   string `form:"booking_code"`
	PlayerContact string `form:"player_contact"`
}

type UpdateBookingWaiting struct {
	AccessoriesId *int    `json:"accessories_id"`
	Amount        *int    `json:"amount"`
	Note          *string `json:"note"`
}
