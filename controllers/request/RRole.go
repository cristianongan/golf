package request

import "start/utils"

type AddRoleBody struct {
	PartnerUid  string           `json:"partner_uid" binding:"required"`
	CourseUid   string           `json:"course_uid" binding:"required"`
	Name        string           `json:"name" binding:"required"`
	Description string           `json:"description"`
	Permissions utils.ListString `json:"permissions"`
	Type        string           `json:"type"`
}

type GetListRole struct {
	PageRequest
	Search     string `form:"search"`
	CourseUid  string `form:"course_uid"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	Type       string `form:"type"`
}
