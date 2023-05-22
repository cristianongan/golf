package controllers

import (
	"start/config"
	"start/datasources"
	"start/models"
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
	CustomerName   string `json:"customer_name"`
}

type GetCurrentAppVersionForm struct {
	OsType     string `form:"os_type" binding:"required"`     // IOS, ANDROID
	DeviceType string `form:"device_type" binding:"required"` // PHONE, TABLET
}

type UpdateCurrentAppVersionBody struct {
	Version    string `json:"version" binding:"required"` // X.X.X
	OsType     string `json:"os_type" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"`
	Key        string `json:"key" binding:"required"`
	IsForce    int    `json:"is_force"`
}

/*
Get current app
*/
func (_ *CPublic) GetCurrentAppVersion(c *gin.Context) {
	form := GetCurrentAppVersionForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}
	currentVersion := models.ForceUpdate{
		OsType:     form.OsType,
		DeviceType: form.DeviceType,
	}

	errF := currentVersion.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	okResponse(c, currentVersion)
}

func (_ *CPublic) UpdateCurrentAppVersion(c *gin.Context) {

	body := UpdateCurrentAppVersionBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Key != "cFrdr1Za6qc9kRaOTYEUi18gE5Qeutd5" {
		response_message.BadRequest(c, "Key invalid")
		return
	}

	isValidOs := false
	isValidDeviceType := false

	if body.OsType == "IOS" || body.OsType == "ANDROID" {
		isValidOs = true
	}

	if body.DeviceType == "PHONE" || body.DeviceType == "TABLET" {
		isValidDeviceType = true
	}

	if !isValidOs || !isValidDeviceType {
		response_message.BadRequest(c, "ostype or device type invalid")
		return
	}

	currentVersion := models.ForceUpdate{
		OsType:     body.OsType,
		DeviceType: body.DeviceType,
	}

	errF := currentVersion.FindFirst()
	if errF != nil {
		errC := currentVersion.Create()
		if errC != nil {
			response_message.BadRequest(c, errC.Error())
			return
		}
	}

	currentVersion.Version = body.Version
	currentVersion.IsForce = body.IsForce

	errUdp := currentVersion.Update()
	if errUdp != nil {
		response_message.BadRequest(c, errUdp.Error())
		return
	}

	okResponse(c, currentVersion)
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

	cusName := booking.CustomerName

	bookResp := PublicGetBookingResp{
		PartnerUid:     booking.PartnerUid,
		CourseUid:      booking.CourseUid,
		CheckInCode:    booking.CheckInCode,
		BookingDate:    booking.BookingDate,
		Hole:           booking.Hole,
		TeeTime:        booking.TeeTime,
		CaddieCode:     booking.CaddieInfo.Code,
		GuestStyleName: booking.GuestStyleName,
		CustomerName:   cusName,
	}

	okResponse(c, bookResp)
}
