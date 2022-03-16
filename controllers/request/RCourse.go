package request

type CreateCourseBody struct {
	Name       string `json:"name" binding:"required"`
	Uid        string `json:"uid" binding:"required"`
	PartnerUid string `json:"partner_uid" binding:"required"`
	Status     string `json:"status"`
}

type GetListCourseForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
}

type UpdateCourseBody struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
