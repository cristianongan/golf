package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CBookingWaiting struct{}

func (_ *CBookingWaiting) CreateBookingWaiting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateBookingWaiting
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "Data format invalid!")
		return
	}

	_, errDate := time.Parse("2006-01-02", body.BookingTime)
	if errDate != nil {
		response_message.BadRequest(c, "Booking Date format invalid!")
		return
	}

	bookingCode := strconv.FormatInt(time.Now().Unix(), 10)

	bookingWaiting := model_booking.BookingWaiting{
		PartnerUid:    body.PartnerUid,
		CourseUid:     body.CourseUid,
		BookingCode:   bookingCode,
		BookingTime:   body.BookingTime,
		PlayerName:    body.PlayerName,
		PlayerContact: body.PlayerContact,
		PeopleList:    body.PeopleList,
		Note:          body.Note,
	}

	err := bookingWaiting.Create(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, bookingWaiting)
}

func (_ *CBookingWaiting) GetBookingWaitingList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWaitingForm{}
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

	bookingWaitingRequest := model_booking.BookingWaiting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	if form.PlayerName != "" {
		bookingWaitingRequest.PlayerName = form.PlayerName
	}

	if form.Date != "" {
		bookingWaitingRequest.BookingTime = form.Date
	}

	list, total, err := bookingWaitingRequest.FindList(db, page)

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

func (_ *CBookingWaiting) DeleteBookingWaiting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingIdStr := c.Param("id")
	bookingId, errId := strconv.ParseInt(bookingIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	bookingWaitingRequest := model_booking.BookingWaiting{}
	bookingWaitingRequest.Id = bookingId
	errF := bookingWaitingRequest.FindFirst(db)

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := bookingWaitingRequest.Delete(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CBookingWaiting) UpdateBookingWaiting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddieIdStr := c.Param("id")
	caddieId, errId := strconv.ParseInt(caddieIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.CreateBookingWaiting
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingWaitingRequest := model_booking.BookingWaiting{}
	bookingWaitingRequest.Id = caddieId

	errF := bookingWaitingRequest.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	if body.BookingTime != "" {
		bookingWaitingRequest.BookingTime = body.BookingTime
	}
	if body.PlayerName != "" {
		bookingWaitingRequest.PlayerName = body.PlayerName
	}
	if body.PlayerContact != "" {
		bookingWaitingRequest.PlayerContact = body.PlayerContact
	}
	if len(body.PeopleList) != 0 {
		bookingWaitingRequest.PeopleList = body.PeopleList
	}
	if body.Note != "" {
		bookingWaitingRequest.Note = body.Note
	}

	err := bookingWaitingRequest.Update(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
