package request

type GetCaddieBookingList struct {
	PageRequest
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date"`
	CaddieCode  string `form:"caddie_code"`
	CaddieName  string `form:"caddie_name"`
}

type GetAgencyBookingList struct {
	PageRequest
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date"`
}

type GetCancelBookingList struct {
	PageRequest
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date"`
}
