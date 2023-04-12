package controllers

import (
	"errors"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CParOfHole struct{}

func (_ *CParOfHole) CreateParOfHole(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateParOfHoleBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	parOfHole := models.ParOfHole{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		CourseType: body.CourseType,
		Course:     body.Course,
		Hole:       body.Hole,
		Par:        body.Par,
		Minute:     body.Minute,
	}

	errC := parOfHole.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, parOfHole)
}

func (_ *CParOfHole) GetListParOfHole(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListParOfHoleForm{}
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

	parOfHoleR := models.ParOfHole{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := parOfHoleR.FindList(db, page)
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

func (_ *CParOfHole) UpdateParOfHole(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	parOfHoleIdStr := c.Param("id")
	parOfHoleId, err := strconv.ParseInt(parOfHoleIdStr, 10, 64)
	if err != nil || parOfHoleId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	// validate body
	body := request.UpdateParOfHoleBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	parOfHole := models.ParOfHole{}
	parOfHole.Id = parOfHoleId
	errF := parOfHole.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.Status != "" {
		parOfHole.Status = body.Status
	}
	if body.CourseType != "" {
		parOfHole.CourseType = body.CourseType
	}
	if body.Course != "" {
		parOfHole.Course = body.Course
	}

	parOfHole.Hole = body.Hole
	parOfHole.Par = body.Par
	parOfHole.Minute = body.Minute

	errUdp := parOfHole.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, parOfHole)
}

func (_ *CParOfHole) DeleteParOfHole(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	parOfHoleIdStr := c.Param("id")
	parOfHoleId, err := strconv.ParseInt(parOfHoleIdStr, 10, 64)
	if err != nil || parOfHoleId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	parOfHole := models.ParOfHole{}
	parOfHole.Id = parOfHoleId
	errF := parOfHole.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := parOfHole.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
