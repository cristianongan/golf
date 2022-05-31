package request

type CreateCaddieNoteBody struct {
	CaddieId string `json:"caddie_id" binding:"required"`
	AtDate   int64  `json:"at_date"`
	Type     string `json:"type"`
	Note     string `json:"note"`
}

type GetListCaddieNoteForm struct {
	PageRequest
	From int64 `form:"from"`
	To   int64 `form:"to"`
}

type UpdateCaddieNoteBody struct {
	AtDate *int64  `json:"at_date"`
	Type   *string `json:"type"`
	Note   *string `json:"note"`
}
