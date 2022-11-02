package request

type GetReportDashboardRequestForm struct {
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}
