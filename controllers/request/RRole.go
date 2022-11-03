package request

import "start/utils"

type AddRoleBody struct {
	PartnerUid  string           `json:"partner_uid" binding:"required"`
	CourseUid   string           `json:"course_uid" binding:"required"`
	Name        string           `json:"name" binding:"required"`
	Description string           `json:"description"`
	Permissions utils.ListString `json:"permissions"`
}

type GetListRole struct {
	PageRequest
	Search     string `form:"search"`
	CourseUid  string `form:"course_uid"`
	PartnerUid string `json:"partner_uid" binding:"required"`
}
