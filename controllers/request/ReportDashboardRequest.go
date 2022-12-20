package request

type GetReportDashboardRequestForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
}

type GetReportRevenueDashboardRequestForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	// Year       string `form:"year" binding:"required"`
}

type GetReportTop10MemberForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	TypeMember string `form:"type_member" binding:"required"`
	TypeDate   string `form:"type_date" binding:"required"`
	Date       string `form:"date" binding:"required"`
}
