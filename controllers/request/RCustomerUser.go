package request

type GetListCustomerUserForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}
