package controllers

import (
	"log"
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

	if body.NumBook <= 0 {
		body.NumBook = 1
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
		CourseUid:  body.CourseCode,
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

	if body.Tee == "" {
		body.Tee = "1"
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	// Find course
	course := models.Course{}
	course.Uid = body.CourseCode
	errFCourse := course.FindFirst(db)
	if errFCourse != nil {
		dataRes.Result.Status = 500
		dataRes.Result.Infor = "Not found course"
		c.JSON(500, dataRes)
		return
	}

	// Check tee time status
	// Check TeeTime Index
	bookTeaTimeIndex := model_booking.Booking{
		PartnerUid:  prof.PartnerUid,
		CourseUid:   prof.CourseUid,
		BookingDate: bookDate,
		TeeTime:     body.TeeOffStr,
		TeeType:     body.Tee,
	}

	listIndex := bookTeaTimeIndex.FindTeeTimeIndexAvaible(db)

	if len(listIndex) == 0 {
		//
		dataRes.Result.Status = 500
		dataRes.Result.Infor = "Tee is full"
		c.JSON(500, dataRes)
		return
	}

	if len(listIndex) > 0 && len(listIndex) < body.NumBook {
		//
		dataRes.Result.Status = 500
		dataRes.Result.Infor = "Tee khong du"
		c.JSON(500, dataRes)
		return
	}

	// Check agency
	// Find Agency
	agency := models.Agency{
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
		AgencyId:   body.AgentCode,
	}
	errFA := agency.FindFirst(db)
	if errFA != nil {
		dataRes.Result.Status = 500
		dataRes.Result.Infor = "Not found agency"
		c.JSON(500, dataRes)
		return
	}

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

		AgentCode:    body.AgentCode,
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

	for i := 0; i < body.NumBook; i++ {
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
			RowIndex:             &listIndex[i],
			AgencyId:             agency.Id,
		}

		if body.IsMainCourse {
			bodyCreate.CourseType = "A"
		} else {
			bodyCreate.CourseType = "B"
		}

		booking := cBooking.CreateBookingCommon(bodyCreate, c, prof)
		if booking == nil {
			//error
			log.Println("CreateBookingOTA error")
		}
	}

	dataRes.BookID = bookingOta.Id

	okResponse(c, dataRes)
}
