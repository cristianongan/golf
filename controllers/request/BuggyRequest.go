package request

type CreateBuggyBody struct {
	CourseId string `json:"course_id" binding:"required"`
	Number   int    `json:"number" binding:"required"`
	Origin   string `json:"origin"`
	Note     string `json:"note"`
}

type GetListBuggyForm struct {
	PageRequest
	CourseId string `form:"course_id" json:"course_id"`
}

type UpdateBuggyBody struct {
	Origin *string `json:"origin"`
	Note   *string `json:"note"`
}