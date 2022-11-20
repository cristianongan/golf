package request

type CreateCaddieGroupBody struct {
	GroupName string `json:"group_name" validate:"required"`
	GroupCode string `json:"group_code" validate:"required"`
}

type AddCaddieToGroupBody struct {
	GroupCode  string   `json:"group_code" validate:"required"`
	CaddieList []string `json:"caddie_list"`
}

type GetCaddieGroupList struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}

type MoveCaddieToGroupBody struct {
	GroupCode  string   `json:"group_code" validate:"required"`
	CaddieList []string `json:"caddie_list"`
}
