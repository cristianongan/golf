package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
)

type CTodo struct{}

func (_ *CTodo) CreateTodo(c *gin.Context) {
	var body request.CreateTodoBody
	var a = c.BindJSON(&body)
	if a != nil {
		log.Println(a)
		response_message.BadRequest(c, "")
		return
	}

	base := models.Model{
		Status: constants.STATUS_ENABLE,
	}
	todo := models.Todo{
		Model:   base,
		Content: body.Content,
		Done:    false,
	}

	err := todo.Create()
	if err != nil {
		log.Println(err)
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, todo)
}

func (_ *CTodo) GetTodoList(c *gin.Context) {
	form := request.GetListToDoForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	log.Println(form)

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	todoRequest := models.Todo{}

	list, total, err := todoRequest.FindList(page, form.Done)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}

func (_ *CTodo) DeleteTodo(c *gin.Context) {
	var body request.DeleteTodoBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	todoRequest := models.Todo{}
	todoRequest.Uid = body.Uid
	_, errF := todoRequest.FindFirst()

	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	err := todoRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CTodo) UpdateTodo(c *gin.Context) {
	var body request.UpdateTodoBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	todoRequest := models.Todo{}
	todoRequest.Uid = body.Uid

	oldTodo, errF := todoRequest.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	var older = oldTodo[0]

	todoRequest.Uid = body.Uid
	todoRequest.CreatedAt = older.CreatedAt
	todoRequest.Uid = older.Uid
	todoRequest.Content = older.Content
	todoRequest.Status = older.Status
	todoRequest.Done = body.Done

	err := todoRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
