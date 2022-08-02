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
}
