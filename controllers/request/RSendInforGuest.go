package request

type GetListSendInforGuestForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date" binding:"required"`
	Search      string `form:"search"`
}
