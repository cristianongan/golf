package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CBookingSource struct{}

func (_ *CBookingSource) CreateBookingSource(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body model_booking.BookingSource
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	bookingSourceId := strings.ToUpper(body.BookingSourceId)
	bookingSource := model_booking.BookingSource{
		PartnerUid:      body.PartnerUid,
		CourseUid:       body.CourseUid,
		BookingSourceId: bookingSourceId,
	}

	error := bookingSource.FindFirst(db)
	if error == nil {
		response_message.BadRequest(c, "Booking Source Id đã tồn tại!")
		return
	}
	bookingSource = model_booking.BookingSource{
		PartnerUid:        body.PartnerUid,
		CourseUid:         body.CourseUid,
		BookingSourceName: body.BookingSourceName,
		AgencyId:          body.AgencyId,
		IsPart1TeeType:    body.IsPart1TeeType,
		IsPart2TeeType:    body.IsPart2TeeType,
		IsPart3TeeType:    body.IsPart3TeeType,
		NormalDay:         body.NormalDay,
		Weekend:           body.Weekend,
		NumberOfDays:      body.NumberOfDays,
		BookingSourceId:   bookingSourceId,
	}

	if body.Status != "" {
		bookingSource.Status = body.Status
	}

	err := bookingSource.Create(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, bookingSource)
}

func (_ *CBookingSource) GetBookingSourceList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	bookingSourceRequest := model_booking.BookingSource{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	if form.BookingSourceName != "" {
		bookingSourceRequest.BookingSourceName = form.BookingSourceName
	}

	if form.Status != "" {
		bookingSourceRequest.Status = form.Status
	}

	list, total, err := bookingSourceRequest.FindList(db, page)

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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingIdStr := c.Param("id")
	bookingId, errId := strconv.ParseInt(bookingIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	bookingSourceRequest := model_booking.BookingSource{}
	bookingSourceRequest.Id = bookingId
	bookingSourceRequest.PartnerUid = prof.PartnerUid
	bookingSourceRequest.CourseUid = prof.CourseUid
	errF := bookingSourceRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := bookingSourceRequest.Delete(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CBookingSource) UpdateBookingSource(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	bookingSourceRequest.PartnerUid = prof.PartnerUid
	bookingSourceRequest.CourseUid = prof.CourseUid

	errF := bookingSourceRequest.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.Status != "" {
		bookingSourceRequest.Status = body.Status
	}
	if body.BookingSourceName != "" {
		bookingSourceRequest.BookingSourceName = body.BookingSourceName
	}
	if body.IsPart1TeeType != nil {
		bookingSourceRequest.IsPart1TeeType = *body.IsPart1TeeType
	}
	if body.IsPart2TeeType != nil {
		bookingSourceRequest.IsPart2TeeType = *body.IsPart2TeeType
	}
	if body.IsPart3TeeType != nil {
		bookingSourceRequest.IsPart3TeeType = *body.IsPart3TeeType
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
	if body.AgencyId != 0 {
		bookingSourceRequest.AgencyId = body.AgencyId
	}

	err := bookingSourceRequest.Update(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
