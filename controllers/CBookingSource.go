package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBookingSource struct{}

func (_ *CBookingSource) CreateBookingSource(c *gin.Context, prof models.CmsUser) {
	var body model_booking.BookingSource
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	bookingSource := model_booking.BookingSource{
		PartnerUid:        body.PartnerUid,
		CourseUid:         body.CourseUid,
		BookingSourceName: body.BookingSourceName,
		TeeTime:           body.TeeTime,
		NormalDay:         body.NormalDay,
		Weekend:           body.Weekend,
		NumberOfDays:      body.NumberOfDays,
	}

	err := bookingSource.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, bookingSource)
}

func (_ *CBookingSource) GetBookingSourceList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingSource{}
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

	bookingSourceRequest := model_booking.BookingSource{}

	if form.BookingSourceName != "" {
		bookingSourceRequest.BookingSourceName = form.BookingSourceName
	}

	if form.Status != "" {
		bookingSourceRequest.Status = form.Status
	}

	list, total, err := bookingSourceRequest.FindList(page)

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

func (_ *CBookingSource) DeleteBookingSource(c *gin.Context, prof models.CmsUser) {
	bookingIdStr := c.Param("id")
	bookingId, errId := strconv.ParseInt(bookingIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	bookingSourceRequest := model_booking.BookingSource{}
	bookingSourceRequest.Id = bookingId
	errF := bookingSourceRequest.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := bookingSourceRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CBookingSource) UpdateBookingSource(c *gin.Context, prof models.CmsUser) {
	bookingSourceIdStr := c.Param("id")
	bookingSourceId, errId := strconv.ParseInt(bookingSourceIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateBookingSource
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingSourceRequest := model_booking.BookingSource{}
	bookingSourceRequest.Id = bookingSourceId

	errF := bookingSourceRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	if body.BookingSourceName != "" {
		bookingSourceRequest.BookingSourceName = body.BookingSourceName
	}
	if len(body.TeeTime) != 0 {
		bookingSourceRequest.TeeTime = body.TeeTime
	}
	if body.NormalDay != nil {
		bookingSourceRequest.NormalDay = *body.NormalDay
	}
	if body.Weekend != nil {
		bookingSourceRequest.Weekend = *body.Weekend
	}
	if body.NumberOfDays != 0 {
		bookingSourceRequest.NumberOfDays = body.NumberOfDays
	}

	err := bookingSourceRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
