package request

type GetListMemberCardTypeForm struct {
	PageRequest
	Status     string `form:"status"`
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Name       string `form:"name"`
	Type       string `form:"type"`
	GuestStyle string `form:"guest_style"`
}
