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

type GetFeeByHoleForm struct {
	McTypeId int64 `form:"mc_type_id" binding:"required"`
	Hole     int   `form:"hole"`
}
