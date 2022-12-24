package controllers

import (
	"net/http"
	"start/config"
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

	// Find Course
	course := models.Course{}
	course.Uid = body.CourseCode
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		responseOTA.Result.Status = 500
		responseOTA.Result.Infor = "Course Code not found"
		okResponse(c, responseOTA)
		return
	}

	checkToken := course.ApiKey + body.CourseCode + body.Date
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != body.Token {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "token invalid"

		okResponse(c, responseOTA)
		return
	}

	db := datasources.GetDatabase()

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

	timeDate, _ := time.Parse(constants.DATE_FORMAT, body.Date)
	agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnDate(db, timeDate)
	if errFSP == nil && agencySpecialPrice.Id > 0 {
		GreenFee = agencySpecialPrice.GreenFee
		CaddieFee = agencySpecialPrice.CaddieFee
		// BuggyFee = agencySpecialPrice.BuggyFee
	} else {
		golfFee := models.GolfFee{
			GuestStyle: agency.GuestStyle,
			CourseUid:  body.CourseCode,
		}

		fee, _ := golfFee.GetGuestStyleOnTime(db, timeDate)

		// Lấy giá hole
		GreenFee = utils.GetFeeFromListFee(fee.GreenFee, body.Hole)
		CaddieFee = utils.GetFeeFromListFee(fee.CaddieFee, body.Hole)
		// BuggyFee = utils.GetFeeFromListFee(fee.BuggyFee, body.Hole)
	}

	BuggyFee = utils.GetFeeFromListFee(getBuggyFee(agency.GuestStyle), 18)

	courseType := ""
	if body.IsMainCourse {
		courseType = "A"
	} else {
		courseType = "B"
	}
	RTeeTimeList := models.TeeTimeList{
		CourseUid:   body.CourseCode,
		BookingDate: dateFormat,
		CourseType:  courseType,
	}

	teeTimeList, _, _ := RTeeTimeList.FindAllList(db)
	teeTimeListOTA := []response.TeeTimeOTA{}

	// get các teetime đang bị khóa ở redis
	listTeeTimeLockRedis := getTeeTimeLockRedis(body.CourseCode, dateFormat)

	for _, teeTime := range teeTimeList {

		teeOff, _ := time.Parse(constants.DATE_FORMAT_3, body.Date+" "+teeTime.TeeTime)
		teeOffStr := teeOff.Format("2006-01-02T15:04:05")
		teeType := teeTime.TeeType + teeTime.CourseType

		teeTime1 := models.LockTeeTime{
			TeeTime:   teeTime.TeeTime,
			CourseUid: body.CourseCode,
			TeeType:   teeType,
		}

		hasTeeTimeLock1AOnRedis := false
		for _, teeTimeLockRedis := range listTeeTimeLockRedis {
			if teeTimeLockRedis.TeeTime == teeTime1.TeeTime && teeTimeLockRedis.DateTime == teeTime1.DateTime &&
				teeTimeLockRedis.CourseUid == teeTime1.CourseUid && teeTimeLockRedis.TeeType == teeTime1.TeeType {
				hasTeeTimeLock1AOnRedis = true
			}
		}

		if !hasTeeTimeLock1AOnRedis {
			if errFind1 := teeTime1.FindFirst(db); errFind1 != nil {
				tee, _ := strconv.Atoi(teeTime.TeeType)
				teeTime1A := response.TeeTimeOTA{
					TeeOffStr:    teeTime.TeeTime,
					DateStr:      body.Date,
					TeeOff:       teeOffStr,
					Tee:          int64(tee),
					SlotEmpty:    int64(teeTime.SlotEmpty),
					NumBook:      int64(constants.SLOT_TEE_TIME - teeTime.SlotEmpty),
					IsMainCourse: body.IsMainCourse,
					GreenFee:     GreenFee,
					CaddieFee:    CaddieFee,
					BuggyFee:     BuggyFee,
					Holes:        int64(body.Hole),
				}
				teeTimeListOTA = append(teeTimeListOTA, teeTime1A)
			}
		}
	}

	responseOTA.Data = teeTimeListOTA
	responseOTA.NumTeeTime = int64(len(teeTimeListOTA))
	responseOTA.Result.Status = 200
	responseOTA.Result.Infor = "Get data OK" + "(" + strconv.Itoa(len(teeTimeListOTA)) + " tee time)" + "-at " + body.Date
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

	// Find Course
	course := models.Course{}
	course.Uid = body.CourseCode
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		responseOTA.Result.Status = 500
		responseOTA.Result.Infor = "Course Code not found"
		okResponse(c, responseOTA)
		return
	}

	checkToken := course.ApiKey + body.DateStr + body.TeeOffStr
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
		CurrentTeeTime: body.TeeOffStr,
	}

	if body.Tee == "1" {
		teeTimeSetting.TeeType = "1A"
	}

	if body.Tee == "10" {
		teeTimeSetting.TeeType = "1B"
	}

	// Create redis key tee time lock

	teeTimeRedisKey := config.GetEnvironmentName() + ":" + body.CourseCode + "_" + dateFormat + "_"
	if body.Tee == "1" {
		teeTimeRedisKey += body.TeeOffStr + "_" + "1A"
	}
	if body.Tee == "10" {
		teeTimeRedisKey += body.TeeOffStr + "_" + "1B"
	}

	key := datasources.GetRedisKeyTeeTimeLock(teeTimeRedisKey)
	_, errRedis := datasources.GetCache(key)

	teeTimeRedis := models.LockTeeTimeObj{
		DateTime:       teeTimeSetting.DateTime,
		CourseUid:      teeTimeSetting.CourseUid,
		TeeTime:        teeTimeSetting.TeeTime,
		CurrentTeeTime: teeTimeSetting.TeeTime,
		TeeType:        teeTimeSetting.TeeType,
		TeeTimeStatus:  constants.TEE_TIME_LOCKED,
	}

	if errRedis != nil {
		valueParse, _ := teeTimeRedis.Value()
		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 5*60); err != nil {
			responseOTA.Result = response.ResultOTA{
				Status: http.StatusInternalServerError,
				Infor:  err.Error(),
			}
		} else {
			responseOTA.Result = response.ResultOTA{
				Status: 200,
				Infor:  body.CourseCode + "- Lock teeTime " + body.TeeOffStr + " " + dateFormat,
			}
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

	// Find course
	course := models.Course{}
	course.Uid = body.CourseCode
	errFCourse := course.FindFirstHaveKey()
	if errFCourse != nil {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "Not found course"
		c.JSON(http.StatusInternalServerError, responseOTA)
		return
	}

	checkToken := course.ApiKey + body.DateStr + body.TeeOffStr
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

	// Find course
	course := models.Course{}
	course.Uid = body.CourseCode
	errFCourse := course.FindFirstHaveKey()
	if errFCourse != nil {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "Not found course"
		c.JSON(http.StatusInternalServerError, responseOTA)
		return
	}

	checkToken := course.ApiKey + body.DateStr + body.TeeOffStr
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
	}

	if body.Tee == "1" {
		lockTeeTime.TeeType = "1A"
	}

	if body.Tee == "10" {
		lockTeeTime.TeeType = "1B"
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
