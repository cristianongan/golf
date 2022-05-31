package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CCaddieNote struct{}

func (_ *CCaddieNote) CreateCaddieNote(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieNoteBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		log.Print("BindJSON CaddieNote error")
		response_message.BadRequest(c, "")
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.Uid = body.CaddieId
	errExist := caddieRequest.FindFirst()

	if errExist != nil {
		response_message.BadRequest(c, "Caddie number did not exist in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	caddieNote := models.CaddieNote{
		ModelId:  base,
		CaddieId: caddieRequest.Uid,
		Type:     body.Type,
		Note:     body.Note,
	}

	err := caddieNote.Create()
	if err != nil {
		log.Print("Create caddieNote error")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieNote)
}

func (_ *CCaddieNote) GetCaddieNoteList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListCaddieNoteForm{}
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

	caddieNoteRequest := models.CaddieNote{}

	list, total, err := caddieNoteRequest.FindList(page, form.From, form.To)

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

func (_ *CCaddieNote) DeleteCaddieNote(c *gin.Context, prof models.CmsUser) {
	caddieNoteIdStr := c.Param("id")
	caddieNoteId, errId := strconv.ParseInt(caddieNoteIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	caddieNoteRequest := models.CaddieNote{}
	caddieNoteRequest.Id = caddieNoteId
	errF := caddieNoteRequest.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieNoteRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieNote) UpdateCaddieNote(c *gin.Context, prof models.CmsUser) {
	caddieNoteIdStr := c.Param("id")
	caddieNoteId, errId := strconv.ParseInt(caddieNoteIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateCaddieNoteBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddieNoteRequest := models.CaddieNote{}
	caddieNoteRequest.Id = caddieNoteId

	errF := caddieNoteRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	if body.AtDate != nil {
		caddieNoteRequest.AtDate = *body.AtDate
	}
	if body.Type != nil {
		caddieNoteRequest.Type = *body.Type
	}
	if body.Note != nil {
		caddieNoteRequest.Note = *body.Note
	}

	err := caddieNoteRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
