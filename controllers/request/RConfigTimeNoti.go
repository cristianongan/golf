package request

type GetConfigTimeNoti struct {
	PageRequest
	Status     string `form:"status" json:"status"`
	PartnerUid string `form:"partner_uid" json:"partner_uid"`
	CourseUid  string `form:"course_uid" json:"course_uid"`
}
