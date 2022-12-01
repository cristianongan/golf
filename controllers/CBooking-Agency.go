package controllers

import (
	"errors"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

/*
Get chi tiết Golf Fee của bag: Round, Sub bag
*/
func (_ *CBooking) GetListAgencyCancelBooking(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBookingWithSelectForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.BookingDate == "" {
		response_message.BadRequest(c, errors.New("Chưa chọn ngày").Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	booking := model_booking.Booking{}
	booking.PartnerUid = form.PartnerUid
	booking.CourseUid = form.CourseUid
	booking.BookingDate = form.BookingDate
	booking.BookingCode = form.BookingCode

	list, total, err := booking.FindAgencyCancelBooking(db, page)

	res := response.PageResponse{}

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res = response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}
