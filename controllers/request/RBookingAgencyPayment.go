package request

type GetAllBookingAgencyPayment struct {
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	BookingUid string `form:"booking_uid"`
}
