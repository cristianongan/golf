package request

import "start/utils"

type CreateBookingWaiting struct {
	PartnerUid    string           `json:"partner_uid"`
	CourseUid     string           `json:"course_uid"`
	BookingTime   string           `json:"booking_time"`
	PlayerName    string           `json:"player_name"`
	PlayerContact string           `json:"player_contact"`
	PeopleList    utils.ListString `json:"people_list"`
	Note          string           `json:"note"`
}

type GetListBookingWaitingForm struct {
	PageRequest
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	Date       string `form:"date"`
	PlayerName string `form:"player_name"`
}

type UpdateBookingWaiting struct {
	AccessoriesId *int    `json:"accessories_id"`
	Amount        *int    `json:"amount"`
	Note          *string `json:"note"`
}
