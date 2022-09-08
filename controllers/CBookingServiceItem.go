package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBookingServiceItem struct{}

func (_ *CBookingServiceItem) GetBookingServiceItemList(c *gin.Context, prof models.CmsUser) {
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
	}

	list, total, err := bookingServiceItem.FindList(page)

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
	booking := model_booking.Booking{}
	booking.Uid = body.BookingUid
	errFB := booking.FindFirst()
	if errFB != nil {
		response_message.InternalServerError(c, errFB.Error())
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
	serviceItem.PlayerName = booking.CustomerName

	errC := serviceItem.Create()
	if errC != nil {
		response_message.BadRequest(c, errC.Error())
		return
	}

	//Update lại giá trong booking
	updatePriceWithServiceItem(booking, prof)

	okResponse(c, serviceItem)
}

func (_ *CBookingServiceItem) UdpBookingServiceItemToBag(c *gin.Context, prof models.CmsUser) {
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
	errF := serviceItem.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	serviceItem.DiscountValue = body.DiscountValue
	serviceItem.Amount = body.Amount
	serviceItem.Input = body.Input

	errUdp := serviceItem.Update()

	if errUdp != nil {
		response_message.BadRequest(c, errUdp.Error())
		return
	}

	okResponse(c, serviceItem)
}

func (_ *CBookingServiceItem) DelBookingServiceItemToBag(c *gin.Context, prof models.CmsUser) {
	bookingServiceIdStr := c.Param("id")
	bookingServiceId, err := strconv.ParseInt(bookingServiceIdStr, 10, 64)
	if err != nil || bookingServiceId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	serviceItem := model_booking.BookingServiceItem{}
	serviceItem.Id = bookingServiceId
	errF := serviceItem.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	errDel := serviceItem.Delete()

	if errDel != nil {
		response_message.BadRequest(c, errDel.Error())
		return
	}

	//Find Booking
	booking := model_booking.Booking{}
	booking.Uid = serviceItem.BookingUid

	errFB := booking.FindFirst()
	if errFB != nil {
		log.Println("DelBookingServiceItemToBag", errFB.Error())
	} else {
		//Update lại giá trong booking
		updatePriceWithServiceItem(booking, prof)
	}

	okRes(c)
}
