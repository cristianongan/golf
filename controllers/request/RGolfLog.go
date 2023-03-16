package request

type GetOperationLogForm struct {
	PageRequest
	Status      string `form:"status"`
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	Function    string `form:"function"`
	Module      string `form:"module"`
	Action      string `form:"action"`
	Bag         string `form:"bag"`
	BookingDate string `form:"booking_date"`
}
