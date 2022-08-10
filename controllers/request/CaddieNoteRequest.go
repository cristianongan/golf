package request

type CreateCaddieNoteBody struct {
	CourseUid  string `json:"course_uid"`
	PartnerUid string `json:"partner_uid"`
	CaddieId   int64  `json:"caddie_id" binding:"required"`
	AtDate     int64  `json:"at_date"`
	Type       string `json:"type"`
	Note       string `json:"note"`
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
