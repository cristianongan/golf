package request

type CreateCourseBody struct {
	Name       string  `json:"name" binding:"required"`
	Uid        string  `json:"uid" binding:"required"`
	PartnerUid string  `json:"partner_uid" binding:"required"`
	Status     string  `json:"status"`
	Hole       int     `json:"hole"`
	Address    string  `json:"address"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Icon       string  `json:"icon"`
}

type GetListCourseForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
}

type UpdateCourseBody struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
