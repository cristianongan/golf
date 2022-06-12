package go_controllers

import (
	"encoding/json"
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

	// Get booking
	booking := model_booking.Booking{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.BookingDate,
		Bag:         body.Bag,
	}

	err := booking.FindFirst()

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//Check caddie
	caddie := models.Caddie{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Code:       body.CaddieCode,
	}
	errFC := caddie.FindFirst()
	if errFC != nil {
		response_message.BadRequest(c, "Caddie err "+errFC.Error())
		return
	}
	//Check buggy
	buggy := models.Buggy{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Code:       body.BuggyCode,
	}
	errFB := buggy.FindFirst()
	if errFC != nil {
		response_message.BadRequest(c, "Buggy err "+errFB.Error())
		return
	}

	//Caddie
	caddieInfo := model_booking.BookingCaddie{}
	caddieData, _ := json.Marshal(caddie)
	json.Unmarshal(caddieData, &caddieInfo)
	booking.CaddieId = caddie.Id
	booking.CaddieInfo = caddieInfo
	booking.CaddieStatus = constants.BOOKING_CADDIE_STATUS_IN

	//Buggy
	buggyInfo := model_booking.BookingBuggy{}
	buggyData, _ := json.Marshal(buggy)
	json.Unmarshal(buggyData, &buggyInfo)
	booking.BuggyId = buggy.Id
	booking.BuggyInfo = buggyInfo

	errUdp := booking.Update()
	if err != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	controllers.OkResponse(c, booking)
}

/*
	Add Caddie list
*/
func (_ *CCourseOperating) AddListCaddieBuggyToBooking(c *gin.Context, prof models.CmsUser) {

	body := request.AddCaddieBuggyToBooking{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		controllers.BadRequest(c, bindErr.Error())
		return
	}

}
