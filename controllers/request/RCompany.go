package request

type GetListCompanyForm struct {
	PageRequest
	Status        string `form:"status"`
	PartnerUid    string `form:"partner_uid"`
	CourseUid     string `form:"course_uid"`
	Name          string `form:"name"`
	CompanyTypeId int64  `form:"company_type_id"`
	Phone         string `form:"phone"`
	Code          string `form:"code"`
}
