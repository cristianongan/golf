package request

type CreateCaddieAbsentBody struct {
	CourseId string `json:"course_id" binding:"required"`
	CaddieId string `json:"caddie_id" binding:"required"`
	From     int64  `json:"from" binding:"required"`
	To       int64  `json:"to" binding:"required"`
	Type     string `json:"type"`
	Note     string `json:"note"`
}

type GetListCaddieAbsentForm struct {
	PageRequest
	CourseId string `form:"course_id" json:"course_id"`
	CaddieId string `form:"caddie_id" json:"caddie_id"`
	From     int64  `form:"from"`
	To       int64  `form:"to"`
}

type UpdateCaddieAbsentBody struct {
	From *int64  `json:"from"`
	To   *int64  `json:"to"`
	Type *string `json:"type"`
	Note *string `json:"note"`
}
