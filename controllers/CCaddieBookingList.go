package controllers

import (
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"
	"time"

	"github.com/gin-gonic/gin"
)

type CCaddieBookingList struct{}

func (_ *CCaddieBookingList) GetCaddieBookingList(c *gin.Context, prof models.CmsUser) {
	query := request.GetCaddieBookingList{}
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

	bookingDate, _ := time.Parse("2006-01-02", query.BookingDate)

	bookings := model_booking.BookingList{}

	bookings.BookingDate = bookingDate.Format("02/01/2006")
	bookings.CaddieName = query.CaddieName
	bookings.CaddieCode = query.CaddieCode

	db, total, err := bookings.FindBookingListWithSelect(page)

	var list []response.CaddieBookingResponse
	db.Find(&list)

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

func (_ *CCaddieBookingList) GetAgencyBookingList(c *gin.Context, prof models.CmsUser) {
	query := request.GetAgencyBookingList{}
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

	bookingDate, _ := time.Parse("2006-01-02", query.BookingDate)

	bookings := model_booking.BookingList{}

	bookings.BookingDate = bookingDate.Format("02/01/2006")
	bookings.IsAgency = "0"

	db, total, err := bookings.FindBookingListWithSelect(page)

	var list []response.CaddieAgencyBookingResponse
	db.Find(&list)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}

func (_ *CCaddieBookingList) GetCancelBookingList(c *gin.Context, prof models.CmsUser) {
	const CANCEL = "CANCEL"

	query := request.GetCancelBookingList{}
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

	bookingDate, _ := time.Parse("2006-01-02", query.BookingDate)

	bookings := model_booking.BookingList{}

	bookings.BookingDate = bookingDate.Format("02/01/2006")
	bookings.Status = CANCEL

	db, total, err := bookings.FindBookingListWithSelect(page)

	var list []response.CaddieCancelBookingResponse
	db.Find(&list)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}
