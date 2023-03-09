package request

type RevenueReportFBForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	FromDate    string `form:"from_date"`
	ToDate      string `form:"to_date"`
	TypeService string `form:"type_service"`
}

type RevenueReportDetailFBBagForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Date       string `form:"date" binding:"required"`
}

type RevenueReportDetailFBForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Date       string `form:"date" binding:"required"`
	Type       string `form:"type"`
	GroupCode  string `form:"group_code"`
}

type RevenueBookingReportDetail struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	FromDate    string `form:"from_date"`
	ToDate      string `form:"to_date"`
	GuestStyle  string `form:"guest_style"`
	Month       int    `form:"month"`
	Year        int    `form:"year"`
	BookingDate string `form:"booking_date"`
	Bag         string `form:"bag"`
}

type GetBookingReportRevenueDetail struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date" binding:"required"`
	GuestStyle  string `form:"guest_style"`
	Bag         string `form:"bag"`
}

type ReportBuggyForGuestStyleForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Month      string `form:"month" binding:"required"`
	Year       string `form:"year" binding:"required"`
}

type ReportBagDaily struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
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

type RevenueReportPOSForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
	ServiceId  string `form:"service_id"`
	Type       string `form:"type"`
}

type AgencyPaymentReportForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date"`
	FromDate    string `form:"from_date"`
	ToDate      string `form:"to_date"`
	AgencyName  string `form:"agency_name"`
}

type UpdateReportBody struct {
	PartnerUid  string `form:"partner_uid" json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid   string `form:"course_uid" json:"course_uid" binding:"required"`   // San Golf
	Bag         string `form:"bag" json:"bag"`
	BillNo      string `form:"bill_no" json:"bill_no"`
	BookingDate string `form:"booking_date" json:"booking_date"`
}
