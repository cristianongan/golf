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

	bookings := model_booking.BookingList{}

	if query.BookingDate != "" {
		bookingDate, _ := time.Parse("2006-01-02", query.BookingDate)
		bookings.BookingDate = bookingDate.Format("02/01/2006")
	}

	if query.FromDate != "" {
		bookings.FromDate = query.FromDate
	}

	if query.ToDate != "" {
		bookings.ToDate = query.ToDate
	}

	bookings.CaddieName = query.CaddieName
	bookings.CaddieCode = query.CaddieCode
	bookings.PartnerUid = prof.PartnerUid
	bookings.CourseUid = prof.CourseUid
	bookings.HasBookCaddie = "1"
	bookings.HasCaddie = "1"

	db, total, err := bookings.FindBookingListWithSelect(page)

	var list []response.CaddieBookingResponse
	db.Find(&list)

	var result map[int64]map[string]interface{}

	result = make(map[int64]map[string]interface{})

	for _, item := range list {
		caddie := models.Caddie{}
		caddie.Id = item.CaddieId

		if err := caddie.FindFirst(); err == nil {
			result[item.CaddieId] = make(map[string]interface{})
			result[item.CaddieId]["caddie_info"] = caddie
			if result[item.CaddieId]["total_booking"] == nil {
				result[item.CaddieId]["total_booking"] = 0
			}
			if result[item.CaddieId]["total_agent_booking"] == nil {
				result[item.CaddieId]["total_agent_booking"] = 0
			}
			if result[item.CaddieId]["total_customer_booking"] == nil {
				result[item.CaddieId]["total_customer_booking"] = 0
			}
			if totalBooking, ok := result[item.CaddieId]["total_booking"].(int); ok {
				result[item.CaddieId]["total_booking"] = totalBooking + 1
			}
			if item.AgencyId != 0 {
				if totalAgencyBooking, ok := result[item.CaddieId]["total_agent_booking"].(int); ok {
					result[item.CaddieId]["total_agent_booking"] = totalAgencyBooking + 1
				}
			} else {
				if totalCustomerBooking, ok := result[item.CaddieId]["total_customer_booking"].(int); ok {
					result[item.CaddieId]["total_customer_booking"] = totalCustomerBooking + 1
				}
			}
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
