package request

type RevenueReportFBForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	FromDate    string `form:"from_date"`
	ToDate      string `form:"to_date"`
	TypeService string `form:"type_service"`
}

type RevenueReportDetailFBForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
	Service    string `form:"service"`
	Name       string `form:"name"`
}

type RevenueBookingReportDetail struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
	GuestStyle string `form:"guest_style"`
	Month      int    `form:"month"`
	Year       int    `form:"year"`
}

type ReportBuggyForGuestStyleForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Month      string `form:"month"`
	Year       string `form:"year"`
}

type ReportBagDaily struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date"`
}

type ReportSalePOSForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Date       string `form:"date" binding:"required"`
	Type       string `form:"type" binding:"required"`
}
type ReportBuggyGoForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Month      string `form:"month" binding:"required"`
	Year       string `form:"year" binding:"required"`
}
