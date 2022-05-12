package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CCaddieAbsent struct{}

func (_ *CCaddieAbsent) CreateCaddieAbsent(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieAbsentBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseId
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.CaddieId = body.CaddieNum
	caddieRequest.CourseId = body.CourseId
	errExist := caddieRequest.FindFirst()

	if errExist != nil || caddieRequest.ModelId.Id < 1 {
		response_message.BadRequest(c, "Caddie number did not exist in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	caddieAbsent := models.CaddieAbsent{
		ModelId:   base,
		CourseId:  body.CourseId,
		CaddieNum: body.CaddieNum,
		From:      body.From,
		To:        body.To,
		Type:      body.Type,
		Note:      body.Note,
	}

	err := caddieAbsent.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieAbsent)
}

func (_ *CCaddieAbsent) GetCaddieAbsentList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListCaddieAbsentForm{}
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

	caddieRequest := models.CaddieAbsent{}

	if form.CourseId != "" {
		caddieRequest.CourseId = form.CourseId
	}

	if form.CaddieNum != "" {
		caddieRequest.CaddieNum = form.CaddieNum
	}

	list, total, err := caddieRequest.FindList(page, form.From, form.To)

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

func (_ *CCaddieAbsent) DeleteCaddieAbsent(c *gin.Context, prof models.CmsUser) {
	caddieIdStr := c.Param("uid")
	caddieId, errId := strconv.ParseInt(caddieIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	caddieRequest := models.CaddieAbsent{}
	caddieRequest.Id = caddieId
	errF := caddieRequest.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieAbsent) UpdateCaddieAbsent(c *gin.Context, prof models.CmsUser) {
	caddieIdStr := c.Param("uid")
	caddieId, errId := strconv.ParseInt(caddieIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateCaddieAbsentBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddieRequest := models.CaddieAbsent{}
	caddieRequest.Id = caddieId

	errF := caddieRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	if body.From != nil {
		caddieRequest.From = *body.From
	}
	if body.To != nil {
		caddieRequest.To = *body.To
	}
	if body.Type != nil {
		caddieRequest.Type = *body.Type
	}
	if body.Note != nil {
		caddieRequest.Note = *body.Note
	}

	err := caddieRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
