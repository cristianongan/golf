package request

type GetListGroupFeeForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Status     string `form:"status"`
}
