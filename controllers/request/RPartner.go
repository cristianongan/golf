package request

type CreatePartnerBody struct {
	Name   string `json:"name" binding:"required"`
	Uid    string `json:"uid" binding:"required"`
	Status string `json:"status"`
}

type GetListPartnerForm struct {
	PageRequest
}

type UpdatePartnerBody struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
