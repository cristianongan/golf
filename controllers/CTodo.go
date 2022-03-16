package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
)

type CTodo struct{}

func (_ *CTodo) CreateTodo(c *gin.Context) {
	var body request.CreateTodoBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
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
	todoUid := c.Param("uid")
	if todoUid == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	todoRequest := models.Todo{}
	todoRequest.Uid = todoUid
	errF := todoRequest.FindFirst()

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
	todoUid := c.Param("uid")
	if todoUid == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	var body request.UpdateTodoBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	todoRequest := models.Todo{}
	todoRequest.Uid = todoUid

	errF := todoRequest.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	todoRequest.Done = body.Done

	err := todoRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
