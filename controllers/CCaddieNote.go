package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CCaddieNote struct{}

func (_ *CCaddieNote) CreateCaddieNote(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateCaddieNoteBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		log.Print("BindJSON CaddieNote error")
		response_message.BadRequest(c, "")
		return
	}

	caddie := models.Caddie{}
	caddie.Id = body.CaddieId
	if err := caddie.FindFirst(db); err != nil {
		response_message.BadRequest(c, "Caddie number did not exist in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	caddieNote := models.CaddieNote{
		ModelId:    base,
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		CaddieId:   body.CaddieId,
		Type:       body.Type,
		Note:       body.Note,
		AtDate:     body.AtDate,
	}

	err := caddieNote.Create(db)
	if err != nil {
		log.Print("Create caddieNote error")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieNote)
}

func (_ *CCaddieNote) GetCaddieNoteList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	list, total, err := caddieNoteRequest.FindList(db, page, form.From, form.To)

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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddieNoteIdStr := c.Param("id")
	caddieNoteId, errId := strconv.ParseInt(caddieNoteIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	caddieNoteRequest := models.CaddieNote{}
	caddieNoteRequest.Id = caddieNoteId
	caddieNoteRequest.PartnerUid = prof.PartnerUid
	caddieNoteRequest.CourseUid = prof.CourseUid
	errF := caddieNoteRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieNoteRequest.Delete(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieNote) UpdateCaddieNote(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	caddieNoteRequest.PartnerUid = prof.PartnerUid
	caddieNoteRequest.CourseUid = prof.CourseUid

	errF := caddieNoteRequest.FindFirst(db)
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

	err := caddieNoteRequest.Update(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
