package request

type CreateCaddieNoteBody struct {
	CourseId string `json:"course_id" binding:"required"`
	CaddieId string `json:"caddie_id" binding:"required"`
	AtDate   int64  `json:"at_date"`
	Type     string `json:"type"`
	Note     string `json:"note"`
}

type GetListCaddieNoteForm struct {
	PageRequest
	CourseId string `form:"course_id"`
	From     int64  `form:"from"`
	To       int64  `form:"to"`
}

type UpdateCaddieNoteBody struct {
	AtDate *int64  `json:"at_date"`
	Type   *string `json:"type"`
	Note   *string `json:"note"`
}
