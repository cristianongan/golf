package request

type CreateTodoBody struct {
	Content string `json:"content" binding:"required"`
}

type GetListToDoForm struct {
	PageRequest
	Done *bool `form:"done"`
}

type DeleteTodoBody struct {
	Uid string `json:"uid" binding:"required"`
}

type UpdateTodoBody struct {
	Uid  string `json:"uid" binding:"required"`
	Done bool   `form:"done" binding:"required"`
}
