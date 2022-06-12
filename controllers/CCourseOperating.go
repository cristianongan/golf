package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_booking "start/models/booking"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CCourseOperating struct{}

/*
 Danh sách booking for caddie on course
 Role: Booking đã checkin, chưa checkout và chưa out Caddies
*/
func (_ *CCourseOperating) GetListBookingCaddieOnCourse(c *gin.Context, prof models.CmsUser) {
	form := request.GetBookingForCaddieOnCourseForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	bookingR := model_booking.Booking{
		PartnerUid:  form.PartnerUid,
		CourseUid:   form.CourseUid,
		BookingDate: form.BookingDate,
	}

	list := bookingR.FindForCaddieOnCourse()

	okResponse(c, list)
}

/*
	Add Caddie short
	Chưa tạo flight
*/
func (_ *CCourseOperating) AddCaddieBuggyToBooking(c *gin.Context, prof models.CmsUser) {
	body := request.AddCaddieBuggyToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.PartnerUid == "" || body.CourseUid == "" || body.BookingDate == "" || body.Bag == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	errB, booking := addCaddieBuggyToBooking(body.PartnerUid, body.CourseUid, body.BookingDate, body.Bag, body.CaddieCode, body.BuggyCode)
	if errB != nil {
		response_message.InternalServerError(c, errB.Error())
		return
	}

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, booking)
}

/*
	Add Caddie list
	Create Flight
*/
func (_ *CCourseOperating) CreateFlight(c *gin.Context, prof models.CmsUser) {
	body := request.CreateFlightBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if len(body.ListData) == 0 {
		response_message.BadRequest(c, "List Data empty")
		return
	}

	// Check các bag ok hết mới tạo flight

	listError := []error{}
	for _, v := range body.ListData {
		errB, _ := addCaddieBuggyToBooking(body.PartnerUid, body.CourseUid, body.BookingDate, v.Bag, v.CaddieCode, v.BuggyCode)
		if errB == nil {
			// errUdp := booking.Update()
			// if errUdp != nil {
			// 	listError = append(listError, errUdp)
			// }
		} else {
			listError = append(listError, errB)
		}
	}

	if len(listError) > 0 {
		errRes := response_message.ErrorResponseDataV2{
			StatusCode:  400,
			ErrorDetail: listError,
		}
		badRequest(c, errRes)
		return
	}

	// Create flight
	// flight := model_gostarter.Flight{
	// 	PartnerUid: body.PartnerUid,
	// 	CourseUid:  body.CourseUid,
	// }

	okRes(c)
}
