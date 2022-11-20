package request

type GetListAnnualFeeForm struct {
	PageRequest
	PartnerUid    string `form:"partner_uid"`
	CourseUid     string `form:"course_uid"`
	Year          int    `form:"year"`
	MemberCardUid string `form:"member_card_uid"`
	CardId        string `form:"card_id"`
}
