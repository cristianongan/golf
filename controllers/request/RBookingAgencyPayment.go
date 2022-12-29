package request

type GetAllBookingAgencyPayment struct {
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingCode string `form:"booking_code"`
	BookingUid  string `form:"booking_uid"`
	AgencyId    int64  `form:"agency_id"`
}
