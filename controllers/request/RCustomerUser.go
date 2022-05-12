package request

type GetListCustomerUserForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	Type        string `form:"type"`
	CustomerUid string `json:"customer_uid"`
	Name        string `json:"name"`
}
