package request

type CreateBuggyDiaryBody struct {
	CourseId      string `json:"course_id" binding:"required"`
	BuggyNumber   int    `json:"buggy_number" binding:"required"`
	AccessoriesId int    `json:"accessories_id" binding:"required"`
	Amount        int    `json:"amount"`
	Note          string `json:"note"`
	InputUser     string `json:"input_user"`
}

type GetListBuggyDiaryForm struct {
	PageRequest
	CourseId    string `form:"course_id" json:"course_id"`
	BuggyNumber *int   `form:"buggy_number" json:"buggy_number"`
	From        int64  `form:"from"`
	To          int64  `form:"to"`
}

type UpdateBuggyDiaryBody struct {
	AccessoriesId *int    `json:"accessories_id"`
	Amount        *int    `json:"amount"`
	Note          *string `json:"note"`
}
