package controllers

import (
	"start/config"
	"start/datasources"
	model_booking "start/models/booking"
	"start/utils"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CPublic struct{}

type PublicGetBookingBody struct {
	CheckSum    string `json:"check_sum" binding:"required"`
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid" binding:"required"`
	CheckInCode string `json:"check_in_code" binding:"required"`
	Date        string `json:"date" binding:"required"`
}

type PublicGetBookingResp struct {
	PartnerUid     string `json:"partner_uid"`
	CourseUid      string `json:"course_uid"`
	CheckInCode    string `json:"check_in_code"`
	BookingDate    string `json:"booking_date"`
	Hole           int    `json:"hole"`
	TeeTime        string `json:"tee_time"`
	CaddieCode     string `json:"caddie_code"`
	GuestStyleName string `json:"guest_style_name"`
}

/*
create single payment and
*/
func (_ *CPublic) GetBookingInfo(c *gin.Context) {

	body := PublicGetBookingBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	// Check check_sum
	checkSumMessage := config.GetPassSecretKey() + "|" + body.PartnerUid + "|" + body.CourseUid + "|" + body.CheckInCode + "|" + body.Date
	checkSum := utils.GetSHA256Hash(checkSumMessage)

	if checkSum != body.CheckSum {
		response_message.BadRequest(c, "checksum invalid")
		return
	}

	db := datasources.GetDatabaseWithPartner(body.PartnerUid)

	// Check booking
	booking := model_booking.Booking{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		BookingDate: body.Date,
	}
	booking.CheckInCode = body.CheckInCode

	errF := booking.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	bookResp := PublicGetBookingResp{
		PartnerUid:     booking.PartnerUid,
		CourseUid:      booking.CourseUid,
		CheckInCode:    booking.CheckInCode,
		BookingDate:    booking.BookingDate,
		Hole:           booking.Hole,
		TeeTime:        booking.TeeTime,
		CaddieCode:     booking.CaddieInfo.Code,
		GuestStyleName: booking.GuestStyleName,
	}

	okResponse(c, bookResp)
}
