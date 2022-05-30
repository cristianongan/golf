package request

type GetListAgencyForm struct {
	PageRequest
	Status     string `form:"status"`
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Name       string `form:"name"`
	AgencyId   string `form:"agency_id"`
}
