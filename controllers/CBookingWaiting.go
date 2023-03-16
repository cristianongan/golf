package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
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

	_, errDate := time.Parse(constants.DATE_FORMAT_1, body.BookingTime)
	if errDate != nil {
		response_message.BadRequest(c, "Booking Date format invalid!")
		return
	}

	bookingCode := strconv.FormatInt(utils.GetTimeNow().Unix(), 10)

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

	opLog := models.OperationLog{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_WAITTING_LIST,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: bookingWaiting},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: body.BookingTime,
	}
	go createOperationLog(opLog)

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

	if form.PlayerContact != "" {
		bookingWaitingRequest.PlayerContact = form.PlayerContact
	}

	if form.BookingCode != "" {
		bookingWaitingRequest.BookingCode = form.BookingCode
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
	bookingWaitingRequest.PartnerUid = prof.PartnerUid
	bookingWaitingRequest.CourseUid = prof.CourseUid
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

	opLog := models.OperationLog{
		PartnerUid:  prof.PartnerUid,
		CourseUid:   prof.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_WAITTING_LIST,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Body:        models.JsonDataLog{Data: bookingIdStr},
		ValueOld:    models.JsonDataLog{Data: bookingWaitingRequest},
		ValueNew:    models.JsonDataLog{},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: bookingWaitingRequest.BookingTime,
	}
	go createOperationLog(opLog)

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
	bookingWaitingRequest.PartnerUid = prof.PartnerUid
	bookingWaitingRequest.CourseUid = prof.CourseUid

	errF := bookingWaitingRequest.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	// Data old
	oldData := bookingWaitingRequest

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

	opLog := models.OperationLog{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Function:    constants.OP_LOG_FUNCTION_WAITTING_LIST,
		Action:      constants.OP_LOG_ACTION_UPDATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldData},
		ValueNew:    models.JsonDataLog{Data: bookingWaitingRequest},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: bookingWaitingRequest.BookingTime,
	}
	go createOperationLog(opLog)

	okRes(c)
}
