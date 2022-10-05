package controllers

import (
	"net/http"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// type CBookingOTA struct{}

/*
Booking OTA
*/
func (cBooking *CBooking) CreateBookingOTA(c *gin.Context) {
	dataRes := response.BookingOTARes{}
	bookResult := response.ResultOTA{}
	dataRes.Result = bookResult

	body := request.CreateBookingOTABody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = bindErr.Error()
		c.JSON(500, dataRes)
		return
	}

	// Check token
	checkToken := "CHILINH_TEST" + body.DateStr + body.TeeOffStr + body.BookingCode
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "token invalid"
		dataRes.CourseCode = body.CourseCode

		okResponse(c, dataRes)
		return
	}

	prof := models.CmsUser{
		PartnerUid: "CHI-LINH",
		CourseUid:  "CHI-LINH-01",
		UserName:   "ota",
	}

	//convert booking date
	bookDate, errBD := utils.GetBookingTimeFrom(body.DateStr)
	if errBD != nil {
		dataRes.Result.Status = 500
		dataRes.Result.Infor = errBD.Error()
		c.JSON(500, dataRes)
		return
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	// Check tee time status

	bookingOta := model_booking.BookingOta{
		PartnerUid:   prof.PartnerUid,
		CourseUid:    prof.CourseUid,
		PlayerName:   body.PlayerName,
		Contact:      body.Contact,
		Note:         body.Note,
		NumBook:      body.NumBook,
		Holes:        body.Holes,
		IsMainCourse: body.IsMainCourse,
		Tee:          body.Tee,
		TeeOffStr:    body.TeeOffStr,
		BookAgent:    body.BookAgent,
		GuestStyle:   body.GuestStyle,
		BookingCode:  body.BookingCode,
		EmailConfirm: body.EmailConfirm,

		CaddieFee: body.CaddieFee,
		BuggyFee:  body.BuggyFee,
		GreenFee:  body.GreenFee,
	}

	errCBO := bookingOta.Create(db)
	if errCBO != nil {
		dataRes.Result.Status = 500
		dataRes.Result.Infor = errCBO.Error()
		c.JSON(500, dataRes)
		return
	}

	bodyCreate := request.CreateBookingBody{
		PartnerUid:           prof.PartnerUid,
		CourseUid:            prof.CourseUid,
		BookingDate:          bookDate,
		Hole:                 body.Holes,
		CustomerName:         body.PlayerName,
		CustomerBookingName:  body.Contact,
		CustomerBookingPhone: body.Contact,
		NoteOfBooking:        body.Note,
		TeeTime:              body.TeeOffStr,
		GuestStyle:           body.GuestStyle,
		TeeType:              body.Tee,
		BookingOtaId:         bookingOta.Id,
	}

	if body.IsMainCourse {
		bodyCreate.CourseType = "A"
	} else {
		bodyCreate.CourseType = "B"
	}

	booking := cBooking.CreateBookingCommon(bodyCreate, c, prof)
	if booking == nil {
		return
	}

	dataRes.BookID = bookingOta.Id

	okResponse(c, dataRes)
}
