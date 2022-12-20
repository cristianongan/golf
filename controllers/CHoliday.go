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

type CHoliday struct{}

func (_ *CHoliday) GetListHoliday(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListHolidayForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	holidayR := models.Holiday{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		Year:       form.Year,
	}

	list, _, err := holidayR.FindList(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, list)
}

func (_ *CHoliday) CreateHoliday(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.Holiday{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, "data not valid")
		return
	}

	holiday := models.Holiday{}
	holiday.PartnerUid = body.PartnerUid
	holiday.CourseUid = body.CourseUid
	holiday.Name = body.Name
	holiday.Note = body.Note
	holiday.Year = body.Year
	holiday.From = body.From
	holiday.To = body.To

	errC := holiday.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, holiday)
}

func (_ *CHoliday) UpdateHoliday(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	holidayIdStr := c.Param("id")
	holidayId, err := strconv.ParseInt(holidayIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && holidayId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	holiday := models.Holiday{}
	holiday.Id = holidayId
	holiday.PartnerUid = prof.PartnerUid
	holiday.CourseUid = prof.CourseUid
	errF := holiday.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.CreateHolidayForm{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		holiday.Name = body.Name
	}
	if body.From != "" {
		holiday.From = body.From
	}
	if body.To != "" {
		holiday.To = body.To
	}
	if body.Note != "" {
		holiday.Note = body.Note
	}
	if body.Year != "" {
		holiday.Year = body.Year
	}

	errUdp := holiday.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, holiday)
}

func (_ *CHoliday) DeleteHoliday(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	holidayIdStr := c.Param("id")
	holidayId, err := strconv.ParseInt(holidayIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && holidayId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	holiday := models.Holiday{}
	holiday.Id = holidayId
	holiday.PartnerUid = prof.PartnerUid
	holiday.CourseUid = prof.CourseUid
	errF := holiday.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := holiday.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
