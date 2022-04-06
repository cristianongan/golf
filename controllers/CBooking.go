package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CBooking struct{}

/// --------- Booking ----------
func (_ *CBooking) CreateBooking(c *gin.Context, prof models.CmsUser) {
	body := request.CreateBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	booking := model_booking.Booking{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		TeeType:    body.TeeType,
		TeePath:    body.TeePath,
		TeeTime:    body.TeeTime,
		TeeOffTime: body.TeeTime,
		TurnTime:   body.TurnTime,
		RowIndex:   body.RowIndex,
	}

	if booking.IsDuplicated() {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	// Check xem booking guest hay booking member
	if body.MemberCardUid != "" {
		// Get Member Card
		memberCard := models.MemberCard{}
		memberCard.Uid = body.MemberCardUid
		errFind := memberCard.FindFirst()
		if errFind != nil {
			response_message.BadRequest(c, errFind.Error())
			return
		}

		// Get Owner
		owner, errOwner := memberCard.GetOwner()
		if errOwner != nil {
			response_message.BadRequest(c, errOwner.Error())
			return
		}

		booking.CustomerName = owner.Name

	} else {

	}

	errC := booking.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, booking)
}

func (_ *CBooking) GetListBooking(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBookingForm{}
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

	bookingR := model_booking.Booking{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := bookingR.FindList(page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}

func (_ *CBooking) UpdateBooking(c *gin.Context, prof models.CmsUser) {
	bookingIdStr := c.Param("uid")
	//bookingId, err := strconv.ParseInt(bookingIdStr, 10, 64)
	if bookingIdStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	booking := model_booking.Booking{}
	booking.Uid = bookingIdStr
	errF := booking.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_booking.Booking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Bag != "" {
		booking.Bag = body.Bag
	}
	booking.GuestStyle = body.GuestStyle

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}
