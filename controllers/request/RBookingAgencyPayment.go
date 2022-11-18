package request

type GetAllBookingAgencyPayment struct {
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	BookingCode string `form:"booking_code"`
	BookingUid  string `form:"booking_uid"`
	AgencyId    int64  `form:"agency_id"`
}
