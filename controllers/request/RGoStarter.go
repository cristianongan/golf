package request

type GetBookingForCaddieOnCourseForm struct {
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date"`
}

// Add Caddie, Buggy To Booking
type AddCaddieBuggyToBooking struct {
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	Bag         string `json:"bag"`
	CaddieCode  string `json:"caddie_code"`
	BuggyCode   string `json:"buggy_code"`
	BookingDate string `json:"booking_date"`
}

type AddListCaddieBuggyToBooking struct {
	PartnerUid  string                 `json:"partner_uid"`
	CourseUid   string                 `json:"course_uid"`
	BookingDate string                 `json:"booking_date"`
	ListData    []CaddieBuggyToBooking `json:"list_data"`
	Note        string                 `json:"note"`
}

type CaddieBuggyToBooking struct {
	Bag        string `json:"bag"`
	CaddieCode string `json:"caddie_code"`
	BuggyCode  string `json:"buggy_code"`
}
