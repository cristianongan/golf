package request

type GetListMemberCardForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	OwnerUid   string `form:"owner_uid"`
	Type       string `form:"type"` // BaseType
	McType     string `form:"mc_type"`
	McTypeId   int64  `form:"mc_type_id"`
}
