package request

type GetListBagNoteForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	GolfBag     string `form:"golf_bag"`
	BookingDate string `form:"booking_date"`
}
