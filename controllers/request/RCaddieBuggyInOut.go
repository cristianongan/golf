package request

type GetListCaddieBuggyInOut struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date"`
	Bag         string `form:"bag"`
	CaddieType  string `form:"caddie_type"`
	BuggyType   string `form:"buggy_type"`
}
