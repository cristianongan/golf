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

type CCaddieWorkingTime struct{}

func (_ *CCaddieWorkingTime) CreateCaddieWorkingTime(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieWorkingTimeBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		log.Print("BindJSON CaddieNote error")
		response_message.BadRequest(c, "")
		return
	}

	caddieRequest := models.Caddie{}
	caddieRequest.CaddieId = body.CaddieId
	errExist := caddieRequest.FindFirst()

	if errExist != nil {
		response_message.BadRequest(c, "Caddie IdentityCard did not exist")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	caddieWorkingTime := models.CaddieWorkingTime{
		ModelId:      base,
		CaddieId:     body.CaddieId,
		CheckInTime:  body.CheckInTime,
		CheckOutTime: body.CheckOutTime,
	}

	err := caddieWorkingTime.Create()
	if err != nil {
		log.Print("Create caddieNote error")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieWorkingTime)
}

func (_ *CCaddieWorkingTime) GetCaddieWorkingTimeDetail(c *gin.Context, prof models.CmsUser) {
	form := request.GetListCaddieWorkingTimeForm{}
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

	caddieWorkingTimeRequest := models.CaddieWorkingTimeResponse{}

	if form.CaddieId != "" {
		caddieWorkingTimeRequest.CaddieId = form.CaddieId
	}

	if form.CaddieName != "" {
		caddieWorkingTimeRequest.CaddieName = form.CaddieName
	}

	list, total, err := caddieWorkingTimeRequest.FindList(page, form.From, form.To)

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

func (_ *CCaddieWorkingTime) DeleteCaddieWorkingTime(c *gin.Context, prof models.CmsUser) {
	idRequest := c.Param("id")
	caddieIdIncrement, errId := strconv.ParseInt(idRequest, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	caddieWorkingTimeRequest := models.CaddieWorkingTime{}
	caddieWorkingTimeRequest.Id = caddieIdIncrement
	errF := caddieWorkingTimeRequest.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := caddieWorkingTimeRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieWorkingTime) UpdateCaddieWorkingTime(c *gin.Context, prof models.CmsUser) {
	idStr := c.Param("id")
	caddieId, errId := strconv.ParseInt(idStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateCaddieWorkingTimeBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	caddieRequest := models.CaddieWorkingTime{}
	caddieRequest.Id = caddieId

	errF := caddieRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.CaddieId != nil {
		caddieRequest.CaddieId = *body.CaddieId
	}
	if body.CheckInTime != nil {
		caddieRequest.CheckOutTime = *body.CheckOutTime
	}
	if body.CheckInTime != nil {
		caddieRequest.CheckInTime = *body.CheckInTime
	}

	err := caddieRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
