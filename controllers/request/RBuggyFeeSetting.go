package request

type GetListBuggyFeeSetting struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	SettingId  int64  `form:"setting_id"`
}
