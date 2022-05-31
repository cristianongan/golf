package request

type GetListAgencyForm struct {
	PageRequest
	Status     string `form:"status"`
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Name       string `form:"name"`
	AgencyId   string `form:"agency_id"`
}

type GetListAgencySpecialPriceForm struct {
	PageRequest
	Status      string `form:"status"`
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	Name        string `form:"name"`
	AgencyId    int64  `form:"agency_id"`
	AgencyIdStr string `form:"agency_id_str"`
}
