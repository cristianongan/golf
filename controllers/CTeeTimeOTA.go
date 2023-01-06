package controllers

import (
	"encoding/json"
	"log"
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

	// Get Setting để tạo list tee time
	cBookingSetting := CBookingSetting{}
	form := request.GetListBookingSettingForm{
		CourseUid: body.CourseCode,
		OnDate:    dateFormat,
	}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	weekday := strconv.Itoa(int(timeDate.Weekday()) + 1)
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

	// get các teetime đang bị khóa ở redis
	listTeeTimeLockRedis := getTeeTimeLockRedis(body.CourseCode, dateFormat, "1A")

	index := 0
	for partIndex, part := range timeParts {
		if !part.IsHideTeePart {
			endTime, _ := utils.ConvertHourToTime(part.EndPart)
			teeTimeInit, _ := utils.ConvertHourToTime(part.StartPart)
			for {
				index += 1

				hour := teeTimeInit.Hour()
				minute := teeTimeInit.Minute()

				hourStr_ := strconv.Itoa(hour)
				if hour < 10 {
					hourStr_ = "0" + hourStr_
				}
				minuteStr := strconv.Itoa(minute)
				if minute < 10 {
					minuteStr = "0" + minuteStr
				}

				hourStr := hourStr_ + ":" + minuteStr

				teeOff, _ := time.Parse(constants.DATE_FORMAT_3, body.Date+" "+hourStr)
				teeOffStr := teeOff.Format("2006-01-02T15:04:05")

				teeTime1 := models.LockTeeTime{
					TeeTime:   hourStr,
					DateTime:  dateFormat,
					CourseUid: body.CourseCode,
					TeeType:   "1A",
				}

				hasTeeTimeLock1AOnRedis := false
				// Get số slot tee time còn trống
				teeTimeSlotEmptyRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty" + "_" + body.CourseCode + "_" + dateFormat + "_" + "1A" + "_" + hourStr
				slotStr, _ := datasources.GetCache(teeTimeSlotEmptyRedisKey)
				slotEmpty, _ := strconv.Atoi(slotStr)

				if slotEmpty == 0 {
					// Check tiếp nếu tee time đó đã bị khóa ở CMS
					for _, teeTimeLockRedis := range listTeeTimeLockRedis {
						if teeTimeLockRedis.TeeTime == teeTime1.TeeTime && teeTimeLockRedis.DateTime == teeTime1.DateTime &&
							teeTimeLockRedis.CourseUid == teeTime1.CourseUid && teeTimeLockRedis.TeeType == teeTime1.TeeType {
							hasTeeTimeLock1AOnRedis = true
							break
						}
					}
				} else {
					if slotEmpty == constants.SLOT_TEE_TIME {
						hasTeeTimeLock1AOnRedis = true
					}
				}

				if !hasTeeTimeLock1AOnRedis {
					teeTime1A := response.TeeTimeOTA{
						TeeOffStr:    hourStr,
						DateStr:      body.Date,
						TeeOff:       teeOffStr,
						Tee:          1,
						Part:         int64(partIndex),
						TimeIndex:    int64(index),
						NumBook:      int64(constants.SLOT_TEE_TIME - slotEmpty),
						IsMainCourse: body.IsMainCourse,
						GreenFee:     GreenFee,
						CaddieFee:    CaddieFee,
						BuggyFee:     BuggyFee,
						Holes:        int64(body.Hole),
					}
					teeTimeList = append(teeTimeList, teeTime1A)
				}

				teeTimeInit = teeTimeInit.Add(time.Minute * time.Duration(bookSetting.TeeMinutes))

				if teeTimeInit.Unix() > endTime.Unix() {
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

	responseOTA := response.LockTeeTimeRes{
		IsMainCourse: body.IsMainCourse,
		Token:        nil,
		CourseCode:   body.CourseCode,
		DateStr:      body.DateStr,
	}

	if body.NumBook <= 0 {
		responseOTA.Result.Status = 500
		responseOTA.Result.Infor = "NumBook invalid!"
		okResponse(c, responseOTA)
		return
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

	// validate slot tee time lock
	teeTimeSetting := models.LockTeeTime{
		DateTime:       dateFormat,
		CourseUid:      body.CourseCode,
		TeeTime:        body.TeeOffStr,
		CurrentTeeTime: body.TeeOffStr,
		ModelId: models.ModelId{
			CreatedAt: time.Now().Unix(),
		},
	}

	if body.Tee == "1" {
		teeTimeSetting.TeeType = "1A"
	}

	if body.Tee == "10" {
		teeTimeSetting.TeeType = "1B"
	}

	teeTimeSlotEmptyRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty" + "_" + teeTimeSetting.CourseUid + "_" + teeTimeSetting.DateTime + "_" + teeTimeSetting.TeeType + "_" + teeTimeSetting.TeeTime
	slotStr, _ := datasources.GetCache(teeTimeSlotEmptyRedisKey)
	slotBook, _ := strconv.Atoi(slotStr)
	slotEmpty := constants.SLOT_TEE_TIME - slotBook

	if body.NumBook > slotEmpty {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "Slot lock invalid!"
		responseOTA.Result.NumBook = int(constants.SLOT_TEE_TIME - slotBook)
		okResponse(c, responseOTA)
		return
	}

	// Create redis key tee time lock
	teeTimeRedisKey := ""
	if body.Tee == "1" {
		teeTimeRedisKey = getKeyTeeTimeLockRedis(dateFormat, body.CourseCode, body.TeeOffStr, "1A")
	}

	if body.Tee == "10" {
		teeTimeRedisKey = getKeyTeeTimeLockRedis(dateFormat, body.CourseCode, body.TeeOffStr, "1B")
	}

	key := datasources.GetRedisKeyTeeTimeLock(teeTimeRedisKey)
	_, errRedis := datasources.GetCache(key)

	teeTimeRedis := models.LockTeeTimeWithSlot{
		DateTime:       teeTimeSetting.DateTime,
		CourseUid:      teeTimeSetting.CourseUid,
		TeeTime:        teeTimeSetting.TeeTime,
		CurrentTeeTime: teeTimeSetting.TeeTime,
		TeeType:        teeTimeSetting.TeeType,
		TeeTimeStatus:  constants.TEE_TIME_LOCKED,
		Slot:           body.NumBook,
		Type:           constants.BOOKING_OTA,
		Note:           "Khóa từ Booking OTA",
		ModelId: models.ModelId{
			CreatedAt: teeTimeSetting.CreatedAt,
		},
	}

	if errRedis != nil {
		valueParse, _ := teeTimeRedis.Value()
		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 0); err != nil {
			responseOTA.Result = response.ResultLockTeeTimeOTA{
				Status: http.StatusInternalServerError,
				Infor:  err.Error(),
			}
		} else {
			responseOTA.Result = response.ResultLockTeeTimeOTA{
				Status: 200,
				Infor:  body.CourseCode + "- Lock teeTime " + body.TeeOffStr + " " + dateFormat,
			}
			datasources.SetCache(teeTimeSlotEmptyRedisKey, slotBook+body.NumBook, 0)
			// Bắn socket để client update ui
			go func() {
				cNotification := CNotification{}
				cNotification.PushNotificationCreateBookingOTA("")
			}()
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

	teeTimeRedisKey := getKeyTeeTimeLockRedis(dateFormat, body.CourseCode, body.TeeOffStr, "1A")
	db := datasources.GetDatabase()
	bookings := model_booking.BookingList{}
	bookings.CourseUid = body.CourseCode
	bookings.BookingDate = dateFormat
	bookings.TeeTime = body.TeeOffStr
	bookings.TeeType = "1"
	bookings.CourseType = "A"

	_, total, _ := bookings.FindAllBookingNotCancelList(db)

	listKey, errRedis := datasources.GetAllKeysWith(teeTimeRedisKey)
	listTeeTimeLockRedis := []models.LockTeeTimeWithSlot{}
	if errRedis == nil && len(listKey) > 0 {
		strData, _ := datasources.GetCaches(listKey...)
		for _, data := range strData {

			byteData := []byte(data.(string))
			teeTime := models.LockTeeTimeWithSlot{}
			err2 := json.Unmarshal(byteData, &teeTime)
			if err2 == nil {
				listTeeTimeLockRedis = append(listTeeTimeLockRedis, teeTime)
			}
		}
	}

	for _, item := range listTeeTimeLockRedis {
		total -= int64(item.Slot)
	}

	err := datasources.DelCacheByKey(teeTimeRedisKey)
	log.Print("runCheckLockTeeTime", err)

	// Bắn socket để client update ui
	go func() {
		cNotification := CNotification{}
		cNotification.PushNotificationCreateBookingOTA("")
	}()

	slotTeeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty" + "_" + body.CourseCode + "_" + dateFormat + "_" + "1A" + "_" + body.TeeOffStr
	if total > 0 {
		if err := datasources.SetCache(slotTeeTimeRedisKey, total, 0); err != nil {
			log.Print("updateSlotTeeTime", err)
		}
	} else {
		err := datasources.DelCacheByKey(slotTeeTimeRedisKey)
		log.Print("runCheckLockTeeTime", err)
	}
	okResponse(c, responseOTA)
}
