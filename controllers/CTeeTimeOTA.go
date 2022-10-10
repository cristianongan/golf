package controllers

import (
	"net/http"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"

	"start/datasources"
	model_booking "start/models/booking"
	"start/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CTeeTimeOTA struct{}

/*
GetTeeTimeList
*/
func (cBooking *CTeeTimeOTA) GetTeeTimeList(c *gin.Context) {
	body := request.GetTeeTimeOTAList{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	bookingDate, _ := time.Parse("2006-01-02", body.Date)
	dateFormat := bookingDate.Format("02/01/2006")

	responseOTA := response.GetTeeTimeOTAResponse{
		IsMainCourse: body.IsMainCourse,
		Token:        nil,
		CourseCode:   body.CourseCode,
		GuestCode:    body.Guest_Code,
		Date:         body.Date,
	}

	checkToken := "CHILINH_TEST" + body.CourseCode + body.Date
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "token invalid"

		okResponse(c, responseOTA)
		return
	}

	db := datasources.GetDatabase()

	// Lấy rate từ Course
	course := models.Course{}
	course.Uid = body.CourseCode
	if errCourse := course.FindFirst(db); errCourse != nil {
		responseOTA.Result.Status = 500
		responseOTA.Result.Infor = "Course Code not found"
		okResponse(c, responseOTA)
		return
	}
	responseOTA.GolfPriceRate = course.RateGolfFee

	// Lấy Fee
	agency := models.Agency{}
	agency.AgencyId = body.OTA_Code
	agency.CourseUid = body.CourseCode
	errFindAgency := agency.FindFirst(db)
	if errFindAgency != nil || agency.Id == 0 {
		responseOTA.Result.Status = 500
		responseOTA.Result.Infor = errFindAgency.Error()
		okResponse(c, responseOTA)
		return
	}
	agencySpecialPriceR := models.AgencySpecialPrice{
		AgencyId:  agency.Id,
		CourseUid: body.CourseCode,
	}

	GreenFee := int64(0)
	CaddieFee := int64(0)
	BuggyFee := int64(0)

	timeDate := time.Unix(utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT, body.Date), 0)
	agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnDate(db, timeDate)
	if errFSP == nil && agencySpecialPrice.Id > 0 {
		GreenFee = agencySpecialPrice.GreenFee
		CaddieFee = agencySpecialPrice.CaddieFee
		BuggyFee = agencySpecialPrice.BuggyFee
	} else {
		golfFee := models.GolfFee{
			GuestStyle: agency.GuestStyle,
			CourseUid:  body.CourseCode,
		}

		fee, _ := golfFee.GetGuestStyleOnTime(db, timeDate)

		// Lấy giá hole 18
		GreenFee = fee.GreenFee[1].Fee
		CaddieFee = fee.CaddieFee[1].Fee
		BuggyFee = fee.BuggyFee[1].Fee
	}

	cBookingSetting := CBookingSetting{}
	form := request.GetListBookingSettingForm{
		CourseUid: body.CourseCode,
	}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	weekday := strconv.Itoa(int(time.Now().Weekday()))
	bookSetting := model_booking.BookingSetting{}

	teeTimeList := []response.TeeTimeOTA{}
	for _, data := range listSettingDetail {
		if strings.ContainsAny(data.Dow, weekday) {
			bookSetting = data
			break
		}
	}

	timeParts := []response.TeeTimePartOTA{
		{
			IsHideTeePart: bookSetting.IsHideTeePart1,
			StartPart:     bookSetting.StartPart1,
			EndPart:       bookSetting.EndPart1,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart2,
			StartPart:     bookSetting.StartPart2,
			EndPart:       bookSetting.EndPart2,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart3,
			StartPart:     bookSetting.StartPart3,
			EndPart:       bookSetting.EndPart3,
		},
	}

	index := 0
	for partIndex, part := range timeParts {
		if !part.IsHideTeePart {
			endTime, _ := utils.ConvertHourToTime(part.EndPart)
			teeTimeInit, _ := utils.ConvertHourToTime(part.StartPart)
			for {
				index += 1
				hourStr := strconv.Itoa(teeTimeInit.Hour()) + ":" + strconv.Itoa(teeTimeInit.Minute())

				teeOff, _ := time.Parse(constants.DATE_FORMAT_3, body.Date+" "+hourStr)
				teeOffStr := teeOff.Format("2006-01-02T15:04:05")

				teeTime1 := models.LockTeeTime{
					TeeTime:   hourStr,
					DateTime:  dateFormat,
					CourseUid: body.CourseCode,
					TeeType:   "1",
				}

				if errFind1 := teeTime1.FindFirst(db); errFind1 != nil {
					teeTime1A := response.TeeTimeOTA{
						TeeOffStr:    hourStr,
						DateStr:      body.Date,
						TeeOff:       teeOffStr,
						Tee:          1,
						Part:         int64(partIndex),
						TimeIndex:    int64(index),
						NumBook:      0,
						IsMainCourse: body.IsMainCourse,
						GreenFee:     GreenFee,
						CaddieFee:    CaddieFee,
						BuggyFee:     BuggyFee,
						Holes:        18,
					}
					teeTimeList = append(teeTimeList, teeTime1A)
				}

				teeTime10 := models.LockTeeTime{
					TeeTime:   hourStr,
					DateTime:  dateFormat,
					CourseUid: body.CourseCode,
					TeeType:   "10",
				}

				if errFind10 := teeTime10.FindFirst(db); errFind10 != nil {
					teeTime10A := response.TeeTimeOTA{
						TeeOffStr:    hourStr,
						DateStr:      body.Date,
						TeeOff:       teeOffStr,
						Tee:          10,
						Part:         int64(partIndex),
						TimeIndex:    int64(index),
						NumBook:      0,
						IsMainCourse: body.IsMainCourse,
						GreenFee:     GreenFee,
						CaddieFee:    CaddieFee,
						BuggyFee:     BuggyFee,
						Holes:        18,
					}
					teeTimeList = append(teeTimeList, teeTime10A)
				}

				teeTimeInit = teeTimeInit.Add(time.Minute * time.Duration(bookSetting.TeeMinutes))

				if teeTimeInit.After(endTime) {
					break
				}
			}
		}
	}

	responseOTA.Data = teeTimeList
	responseOTA.NumTeeTime = int64(len(teeTimeList))
	responseOTA.Result.Status = 200
	responseOTA.Result.Infor = "Get data OK" + "(" + strconv.Itoa(len(teeTimeList)) + " tee time)" + "-at " + body.Date
	okResponse(c, responseOTA)
}

/*
Lock Tee Time
*/
func (cBooking *CTeeTimeOTA) LockTeeTime(c *gin.Context) {
	body := request.RTeeTimeOTA{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	responseOTA := response.TeeTimeStatus{
		IsMainCourse: body.IsMainCourse,
		Token:        nil,
		CourseCode:   body.CourseCode,
		DateStr:      body.DateStr,
	}

	checkToken := "CHILINH_TEST" + body.DateStr + body.TeeOffStr
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "token invalid"

		okResponse(c, responseOTA)
		return
	}

	bookingDate, _ := time.Parse("2006-01-02", body.DateStr)
	dateFormat := bookingDate.Format("02/01/2006")

	teeTimeSetting := models.LockTeeTime{
		DateTime:       dateFormat,
		CourseUid:      body.CourseCode,
		TeeTime:        body.TeeOffStr,
		TeeType:        body.Tee,
		CurrentTeeTime: body.TeeOffStr,
	}
	db := datasources.GetDatabase()
	teeTimeSetting.TeeTimeStatus = constants.TEE_TIME_LOCKED

	errC := teeTimeSetting.Create(db)
	if errC != nil {
		responseOTA.Result = response.ResultOTA{
			Status: http.StatusInternalServerError,
			Infor:  errC.Error(),
		}
	} else {
		responseOTA.Result = response.ResultOTA{
			Status: 200,
			Infor:  body.CourseCode + "- Lock teeTime " + body.TeeOffStr + " " + dateFormat,
		}
	}

	okResponse(c, responseOTA)
}

/*
Tee Time Status
*/
func (cBooking *CTeeTimeOTA) TeeTimeStatus(c *gin.Context) {
	body := request.RTeeTimeOTA{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	db := datasources.GetDatabase()
	responseOTA := response.TeeTimeStatus{
		IsMainCourse: body.IsMainCourse,
		Token:        nil,
		CourseCode:   body.CourseCode,
		DateStr:      body.DateStr,
	}

	checkToken := "CHILINH_TEST" + body.DateStr + body.TeeOffStr
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "token invalid"

		okResponse(c, responseOTA)
		return
	}

	bookingDate, _ := time.Parse("2006-01-02", body.DateStr)
	dateFormat := bookingDate.Format("02/01/2006")

	lockTeeTime := models.LockTeeTime{
		CourseUid: body.CourseCode,
		TeeTime:   body.TeeOffStr,
		DateTime:  dateFormat,
		TeeType:   body.Tee,
	}

	errFind := lockTeeTime.FindFirst(db)
	if errFind == nil && (lockTeeTime.TeeTimeStatus == constants.TEE_TIME_LOCKED) {
		responseOTA.Result = response.ResultOTA{
			Status: http.StatusInternalServerError,
			Infor:  "Tee time is locked!",
		}
	} else {
		bookingList := model_booking.BookingList{
			TeeTime:     body.TeeOffStr,
			BookingDate: dateFormat,
		}
		_, total, _ := bookingList.FindAllBookingList(db)
		responseOTA.Result = response.ResultOTA{
			Status: 200,
			Infor:  body.TeeOffStr + " have " + strconv.Itoa(int(4-total)) + " book is valid",
		}
	}

	okResponse(c, responseOTA)
}

/*
Unlock Tee Time
*/
func (cBooking *CTeeTimeOTA) UnlockTeeTime(c *gin.Context) {
	body := request.RTeeTimeOTA{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}
	responseOTA := response.TeeTimeStatus{
		IsMainCourse: body.IsMainCourse,
		Token:        nil,
		CourseCode:   body.CourseCode,
		DateStr:      body.DateStr,
	}

	checkToken := "CHILINH_TEST" + body.DateStr + body.TeeOffStr
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "token invalid"

		okResponse(c, responseOTA)
		return
	}

	bookingDate, _ := time.Parse("2006-01-02", body.DateStr)
	dateFormat := bookingDate.Format("02/01/2006")
	lockTeeTime := models.LockTeeTime{
		CourseUid: body.CourseCode,
		TeeTime:   body.TeeOffStr,
		DateTime:  dateFormat,
		TeeType:   body.Tee,
	}

	db := datasources.GetDatabase()
	if errFind := lockTeeTime.FindFirst(db); errFind != nil {
		responseOTA.Result = response.ResultOTA{
			Status: http.StatusInternalServerError,
			Infor:  errFind.Error(),
		}
	} else {
		if errFind := lockTeeTime.Delete(db); errFind != nil {
			responseOTA.Result = response.ResultOTA{
				Status: http.StatusInternalServerError,
				Infor:  errFind.Error(),
			}
		} else {
			responseOTA.Result = response.ResultOTA{
				Status: 200,
				Infor:  "Unlock teetime " + body.TeeOffStr + " OK",
			}
		}
	}
	okResponse(c, responseOTA)
}
