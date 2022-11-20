package controllers

import (
	"errors"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBookingCaddyFeeSetting struct{}

func (_ *CBookingCaddyFeeSetting) CreateBookingCaddyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.BookingCaddyFeeSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingCaddyFeeSetting := models.BookingCaddyFeeSetting{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		FromDate:   body.FromDate,
		ToDate:     body.ToDate,
		Fee:        body.Fee,
		Name:       body.Name,
		ModelId:    models.ModelId{Status: body.Status},
	}

	errC := bookingCaddyFeeSetting.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, bookingCaddyFeeSetting)
}

func (_ *CBookingCaddyFeeSetting) GetBookingCaddyFeeSettingList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetBase{}
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

	buggyRequest := models.BookingCaddyFeeSetting{}
	buggyRequest.CourseUid = form.CourseUid
	buggyRequest.PartnerUid = form.PartnerUid

	list, total, err := buggyRequest.FindList(db, page, true)

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

func (_ *CBookingCaddyFeeSetting) DeleteBookingCaddyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || Id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	BookingCaddyFeeSetting := models.BookingCaddyFeeSetting{}
	BookingCaddyFeeSetting.Id = Id
	errF := BookingCaddyFeeSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := BookingCaddyFeeSetting.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
func (_ *CBookingCaddyFeeSetting) UpdateBookingCaddyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || Id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	body := models.BuggyFeeSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingCaddyFeeSetting := models.BookingCaddyFeeSetting{}
	bookingCaddyFeeSetting.Id = Id
	errF := bookingCaddyFeeSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.Status != "" {
		bookingCaddyFeeSetting.Status = body.Status
	}

	errUpd := bookingCaddyFeeSetting.Update(db)
	if errUpd != nil {
		response_message.InternalServerError(c, errUpd.Error())
		return
	}

	okResponse(c, bookingCaddyFeeSetting)
}
