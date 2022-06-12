package go_controllers

import (
	"errors"
	"start/constants"
	"start/controllers"
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

	controllers.OkResponse(c, list)
}

/*
	Add Caddie short
*/
func (_ *CCourseOperating) AddCaddieBuggyToBooking(c *gin.Context, prof models.CmsUser) {
	body := request.AddCaddieBuggyToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		controllers.BadRequest(c, bindErr.Error())
		return
	}

	if body.PartnerUid == "" || body.CourseUid == "" || body.BookingDate == "" || body.Bag == "" {
		controllers.BadRequest(c, errors.New(constants.API_ERR_INVALID_BODY_DATA))
		return
	}

	errB, booking := controllers.AddCaddieBuggyToBooking(body.PartnerUid, body.CourseUid, body.BookingDate, body.Bag, body.CaddieCode, body.BuggyCode)
	if errB != nil {
		response_message.InternalServerError(c, errB.Error())
		return
	}

	errUdp := booking.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	controllers.OkResponse(c, booking)
}

/*
	Add Caddie list
*/
func (_ *CCourseOperating) AddListCaddieBuggyToBooking(c *gin.Context, prof models.CmsUser) {
	body := request.AddListCaddieBuggyToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		controllers.BadRequest(c, bindErr.Error())
		return
	}

	if len(body.ListData) == 0 {
		response_message.BadRequest(c, "List Data empty")
		return
	}

	listError := []error{}

	for _, v := range body.ListData {
		errB, booking := controllers.AddCaddieBuggyToBooking(body.PartnerUid, body.CourseUid, body.BookingDate, v.Bag, v.CaddieCode, v.BuggyCode)
		if errB == nil {
			errUdp := booking.Update()
			if errUdp != nil {
				listError = append(listError, errUdp)
			}
		} else {
			listError = append(listError, errB)
		}
	}

	if len(listError) > 0 {
		controllers.BadRequest(c, listError)
		return
	}

	controllers.OkRes(c)
}
