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
