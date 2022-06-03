package request

type GetListCustomerUserForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	Type        string `form:"type"`
	CustomerUid string `form:"customer_uid"`
	Name        string `form:"name"`
	AgencyId    int64  `form:"agency_id"`
	Phone       string `form:"phone"`
}
