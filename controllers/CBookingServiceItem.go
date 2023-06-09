package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBookingServiceItem struct{}

func (_ *CBookingServiceItem) GetBookingServiceItemList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetBookingServiceItem{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	bookingServiceItem := model_booking.BookingServiceItem{
		GroupCode: query.GroupCode,
		ServiceId: query.ServiceId,
		Name:      query.Name,
		Type:      query.Type,
		ItemCode:  query.ItemCode,
	}
	fromDateUnix := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, query.FromDate)
	toDateUnix := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, query.ToDate)

	list, total, err := bookingServiceItem.FindListWithBooking(db, page, fromDateUnix, toDateUnix)

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

func (_ *CBookingServiceItem) AddBookingServiceItemToBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.AddBookingServiceItem{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.BookingUid == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	//Find booking
	bookingR := model_booking.Booking{}
	bookingR.Uid = body.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)

	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequestFreeMessage(c, "Bag đã CheckOut!")
		return
	}

	itemData, err1 := json.Marshal(body)
	serviceItem := model_booking.BookingServiceItem{}

	if err1 == nil {
		err2 := json.Unmarshal(itemData, &serviceItem)
		if err2 != nil {
			log.Println("AddBookingServiceItemToBag err2", err2.Error())
			response_message.BadRequest(c, err2.Error())
			return
		}
	} else {
		log.Println("AddBookingServiceItemToBag err1", err1.Error())
		response_message.BadRequest(c, err1.Error())
		return
	}

	serviceItem.Bag = booking.Bag
	serviceItem.BillCode = booking.BillCode
	serviceItem.BookingUid = booking.Uid
	serviceItem.PartnerUid = booking.PartnerUid
	serviceItem.CourseUid = booking.CourseUid
	serviceItem.PlayerName = booking.CustomerName
	serviceItem.Location = constants.SERVICE_ITEM_ADD_BY_RECEPTION

	errC := serviceItem.Create(db)
	if errC != nil {
		response_message.BadRequest(c, errC.Error())
		return
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(&booking, prof)

	// Add log
	opLog := models.OperationLog{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_RECEPTION,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: serviceItem},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Bag:         booking.Bag,
		BookingDate: booking.BookingDate,
		BillCode:    booking.BillCode,
		BookingUid:  booking.Uid,
	}

	if serviceItem.Type == constants.RENTAL_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_CHECK_IN
		opLog.Action = constants.OP_LOG_ACTION_ADD_RENTAL
	}

	if serviceItem.Type == constants.DRIVING_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_CHECK_IN
		opLog.Action = constants.OP_LOG_ACTION_ADD_DRIVING
	}

	if serviceItem.Type == constants.PROSHOP_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_CHECK_IN
		opLog.Action = constants.OP_LOG_ACTION_ADD_PROSHOP
	}

	if serviceItem.Type == constants.RESTAURANT_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_CHECK_IN
		opLog.Action = constants.OP_LOG_ACTION_ADD_RESTAURANT
	}

	if serviceItem.Type == constants.KIOSK_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_CHECK_IN
		opLog.Action = constants.OP_LOG_ACTION_ADD_KIOSK
	}

	if serviceItem.Type == constants.MINI_B_SETTING {
		opLog.Function = constants.OP_LOG_FUNCTION_CHECK_IN
		opLog.Action = constants.OP_LOG_ACTION_ADD_MINI_B
	}

	go createOperationLog(opLog)

	okResponse(c, serviceItem)
}

func (_ *CBookingServiceItem) UdpBookingServiceItemToBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingServiceIdStr := c.Param("id")
	bookingServiceId, err := strconv.ParseInt(bookingServiceIdStr, 10, 64)
	if err != nil || bookingServiceId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	body := request.UpdateBookingServiceItem{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	serviceItem := model_booking.BookingServiceItem{}
	serviceItem.Id = bookingServiceId
	serviceItem.PartnerUid = prof.PartnerUid
	serviceItem.CourseUid = prof.CourseUid
	errF := serviceItem.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	serviceItem.DiscountValue = body.DiscountValue
	serviceItem.Amount = body.Amount
	serviceItem.Input = body.Input

	errUdp := serviceItem.Update(db)

	if errUdp != nil {
		response_message.BadRequest(c, errUdp.Error())
		return
	}

	okResponse(c, serviceItem)
}

func (_ *CBookingServiceItem) DelBookingServiceItemToBag(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	bookingServiceIdStr := c.Param("id")
	bookingServiceId, err := strconv.ParseInt(bookingServiceIdStr, 10, 64)
	if err != nil || bookingServiceId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	serviceItem := model_booking.BookingServiceItem{}
	serviceItem.Id = bookingServiceId
	errF := serviceItem.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	bookingR := model_booking.Booking{}
	bookingR.Uid = serviceItem.BookingUid
	booking, errF := bookingR.FindFirstByUId(db)

	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if booking.BagStatus == constants.BAG_STATUS_CHECK_OUT {
		response_message.BadRequestFreeMessage(c, "Bag đã CheckOut!")
		return
	}

	// update service cart

	errDel := serviceItem.Delete(db)
	if errDel != nil {
		response_message.BadRequest(c, errDel.Error())
		return
	}

	// update amout bill
	if serviceItem.ServiceBill > 0 {
		serviceCart := models.ServiceCart{}
		serviceCart.Id = serviceItem.ServiceBill
		if err := serviceCart.FindFirst(db); err == nil {
			serviceCart.Amount -= serviceItem.Amount
			serviceCart.Update(db)
		}
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(&booking, prof)

	okRes(c)
}
