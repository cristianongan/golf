package controllers

import (
	"errors"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CGolfFee struct{}

func (_ *CGolfFee) CreateGolfFee(c *gin.Context, prof models.CmsUser) {
	body := models.GolfFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	// Check Exits
	golfFee := models.GolfFee{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		GuestStyle: body.GuestStyle,
		Dow:        body.Dow,
	}
	errFind := golfFee.FindFirst()
	if errFind == nil || golfFee.Id > 0 {
		response_message.BadRequest(c, "duplicated golf fee")
		return
	}

	golfFee.Status = body.Status
	errC := golfFee.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, golfFee)
}

func (_ *CGolfFee) GetListGolfFee(c *gin.Context, prof models.CmsUser) {
	form := request.GetListGolfFeeForm{}
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

	golfFeeR := models.GolfFee{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	golfFeeR.Status = form.Status
	list, total, err := golfFeeR.FindList(page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CGolfFee) UpdateGolfFee(c *gin.Context, prof models.CmsUser) {
	golfFeeIdStr := c.Param("id")
	golfFeeId, err := strconv.ParseInt(golfFeeIdStr, 10, 64)
	if err != nil || golfFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	golfFee := models.GolfFee{}
	golfFee.Id = golfFeeId
	errF := golfFee.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.GolfFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// if body.Name != "" {
	// 	golfFee.Name = body.Name
	// }

	errUdp := golfFee.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, golfFee)
}

func (_ *CGolfFee) DeleteGolfFee(c *gin.Context, prof models.CmsUser) {
	golfFeeIdStr := c.Param("id")
	golfFeeId, err := strconv.ParseInt(golfFeeIdStr, 10, 64)
	if err != nil || golfFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	golfFee := models.GolfFee{}
	golfFee.Id = golfFeeId
	errF := golfFee.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := golfFee.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
