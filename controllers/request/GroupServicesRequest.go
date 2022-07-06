package request

type CreateGroupServicesBody struct {
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid" binding:"required"`
	GroupCode   string `json:"group_code" binding:"required"`
	GroupName   string `json:"group_name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	DetailGroup string `json:"detail_group"`
}

type GetListGroupServicesForm struct {
	PageRequest
	GroupCode  *string `form:"group_code" json:"group_code"`
	GroupName  *string `form:"group_name" json:"group_name"`
	PartnerUid *string `form:"partner_uid" json:"partner_uid"`
	CourseUid  *string `form:"course_uid" json:"course_uid"`
	Type       *string `form:"type" json:"type"`
}
