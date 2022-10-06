package controllers

import (
	"encoding/json"
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
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = errBD.Error()
		c.JSON(http.StatusInternalServerError, dataRes)
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
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Not found course"
		c.JSON(http.StatusInternalServerError, dataRes)
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
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Tee is full"
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	if len(listIndex) > 0 && len(listIndex) < body.NumBook {
		//
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Tee khong du"
		c.JSON(http.StatusInternalServerError, dataRes)
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
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = "Not found agency"
		c.JSON(http.StatusInternalServerError, dataRes)
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
		dataRes.Result.Status = http.StatusInternalServerError
		dataRes.Result.Infor = errCBO.Error()
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	var errCreateBook error

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
			TeePath:              "MORNING",
		}

		if body.IsMainCourse {
			bodyCreate.CourseType = "A"
		} else {
			bodyCreate.CourseType = "B"
		}

		booking, errBook := cBooking.CreateBookingCommon(bodyCreate, nil, prof)
		if booking == nil {
			//error
			log.Println("CreateBookingOTA error", errBook)
		}
		if errBook != nil {
			errCreateBook = errBook
		}
	}

	if errCreateBook != nil {
		dataRes.Result.Status = 1000
		dataRes.Result.Infor = errCreateBook.Error()
		c.JSON(http.StatusInternalServerError, dataRes)
		return
	}

	bodyByte, _ := body.Marshal()
	_ = json.Unmarshal(bodyByte, &dataRes)

	dataRes.Result.Status = http.StatusOK

	dataRes.BookID = bookingOta.Id

	okResponse(c, dataRes)
}
