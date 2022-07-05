package request

type CreateGroupServicesBody struct {
	GroupCode   string `json:"group_code" binding:"required"`
	GroupName   string `json:"group_name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	DetailGroup string `json:"detail_group"`
}

type GetListGroupServicesForm struct {
	PageRequest
	GroupCode *string `form:"group_code" json:"group_code"`
	GroupName *string `form:"group_name" json:"group_name"`
	Type      *string `form:"type" json:"type"`
}
