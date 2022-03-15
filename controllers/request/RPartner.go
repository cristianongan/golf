package request

type CreatePartnerBody struct {
	Name   string `json:"name" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Status string `json:"status"`
}

type GetListPartnerForm struct {
	PageRequest
}
