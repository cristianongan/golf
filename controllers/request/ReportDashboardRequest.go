package request

type GetReportDashboardRequestForm struct {
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}

type GetReportTop10MemberForm struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	TypeMember string `form:"type_member" binding:"required"`
	TypeDate   string `form:"type_date" binding:"required"`
}
