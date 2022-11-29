package controllers

import (
	"errors"
	"start/constants"
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

	bookings := model_booking.BookingList{}
	bookings.PartnerUid = form.PartnerUid
	bookings.CourseUid = form.CourseUid
	bookings.BookingDate = form.BookingDate
	bookings.GolfBag = form.GolfBag
	bookings.BookingCode = form.BookingCode
	bookings.CaddieCode = form.CaddieCode
	bookings.HasBookCaddie = form.HasBookCaddie
	bookings.CustomerName = form.PlayerName
	bookings.CaddieName = form.CaddieName
	bookings.PlayerOrBag = form.PlayerOrBag
	bookings.IsAgency = "1"
	bookings.BagStatus = constants.BAG_STATUS_CANCEL

	db, total, err := bookings.FindAllBookingList(db)

	res := response.PageResponse{}

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	var list []model_booking.Booking
	db.Debug().Find(&list)
	res = response.PageResponse{
		Total: total,
		Data:  list,
	}

	okResponse(c, res)
}
