package request

type GetListGolfFeeForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Status     string `form:"status"`
}
