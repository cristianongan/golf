package request

import "start/models"

type GetListBuggyFeeSetting struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
}

type CreateBuggyFeeItemSetting struct {
	ParentId int64 `form:"setting_id" binding:"required"`
	models.BuggyFeeItemSetting
}
