package request

type GetListCaddieFeeSettingGroupForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}

type GetListCaddieFeeSettingForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	GroupId    int64  `form:"group_id"`
	OnDate     string `form:"on_date"`
}
