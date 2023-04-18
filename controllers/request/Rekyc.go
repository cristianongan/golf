package request

type EkycGetMemberCardList struct {
	CheckSum   string `json:"check_sum" binding:"required"`
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
}
