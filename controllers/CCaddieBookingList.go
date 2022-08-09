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
	bookings.PartnerUid = prof.PartnerUid
	bookings.CourseUid = prof.CourseUid
	bookings.HasBookCaddie = "1"
	bookings.HasCaddie = "1"

	db, total, err := bookings.FindBookingListWithSelect(page)

	var list []response.CaddieBookingResponse
	db.Find(&list)

	var result map[string]map[string]int

	result = make(map[string]map[string]int)

	for _, item := range list {
		if result[item.CaddieId] == nil {
			result[item.CaddieId] = make(map[string]int)
		}
		result[item.CaddieId]["total_booking"] = result[item.CaddieId]["total_booking"] + 1
		if item.AgencyId != 0 {
			result[item.CaddieId]["total_agent_booking"] = result[item.CaddieId]["total_agent_booking"] + 1
		} else {
			result[item.CaddieId]["total_customer_booking"] = result[item.CaddieId]["total_customer_booking"] + 1
		}
	}

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  result,
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
	bookings.PartnerUid = prof.PartnerUid
	bookings.CourseUid = prof.CourseUid

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
	bookings.PartnerUid = prof.PartnerUid
	bookings.CourseUid = prof.CourseUid

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
