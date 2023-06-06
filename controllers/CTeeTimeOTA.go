package controllers

import (
	"encoding/json"
	"log"
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
Validate TeeType
*/
func validateTeeType(teeType string, listTeeType []models.TeeTypeInfo) bool {
	isOk := false
	for _, v := range listTeeType {
		if v.TeeType == teeType {
			isOk = true
		}

	}
	return isOk
}

/*
GetTeeTimeList
*/
func (cBooking *CTeeTimeOTA) GetTeeTimeList(c *gin.Context) {
	body := request.GetTeeTimeOTAList{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	date, _ := time.Parse("2006-01-02", body.Date)
	bookingDateF := date.Format("02/01/2006")

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

	BuggyFee = utils.GetFeeFromListFee(getBuggyFee(agency.GuestStyle, bookingDateF), 18)

	// Get Setting để tạo list tee time
	cBookingSetting := CBookingSetting{}
	form := request.GetListBookingSettingForm{
		CourseUid: body.CourseCode,
		OnDate:    bookingDateF,
	}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	bookingDateTime, _ := time.Parse(constants.DATE_FORMAT_1, bookingDateF)
	weekday := strconv.Itoa(int(bookingDateTime.Weekday() + 1))
	bookSetting := model_booking.BookingSetting{}

	teeTimeList := []response.TeeTimeOTA{}
	for _, data := range listSettingDetail {
		if strings.ContainsAny(data.Dow, weekday) {
			bookSetting = data
			break
		}
	}

	// Get list Tee Type info
	teeTypeR := models.TeeTypeInfo{
		PartnerUid: course.PartnerUid,
		CourseUid:  course.Uid,
	}
	listTeeType, errLTP := teeTypeR.FindALL()
	if errLTP != nil || len(listTeeType) == 0 {
		log.Println("GetTeeTimeList errLTP error or empty")
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "TeeType Info not yet config"

		okResponse(c, responseOTA)
		return

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

	if body.TeeType != "" && validateTeeType(body.TeeType, listTeeType) {
		// get các teetime đang bị khóa ở redis
		listTeeTimeLockRedis := getTeeTimeLockRedis(body.CourseCode, bookingDateF, body.TeeType)

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
						DateTime:  bookingDateF,
						CourseUid: body.CourseCode,
						TeeType:   body.TeeType,
					}

					hasTeeTimeLock1AOnRedis := false

					// Lấy số slot đã book
					teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(bookingDateF, body.CourseCode, hourStr, body.TeeType)
					rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
					rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
					//

					// Get số slot tee time còn trống
					teeTimeSlotEmptyRedisKey := getKeyTeeTimeSlotRedis(bookingDateF, body.CourseCode, hourStr, body.TeeType)
					slotStr, _ := datasources.GetCache(teeTimeSlotEmptyRedisKey)
					slotLock, _ := strconv.Atoi(slotStr)
					slotEmpty := slotLock + len(rowIndexsRedis)

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
							TeeType:      body.TeeType,
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
	} else {

		for _, teeTypeInfo := range listTeeType {
			teeType := teeTypeInfo.TeeType
			// get các teetime đang bị khóa ở redis
			listTeeTimeLockRedis := getTeeTimeLockRedis(body.CourseCode, bookingDateF, teeType)

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
							DateTime:  bookingDateF,
							CourseUid: body.CourseCode,
							TeeType:   body.TeeType,
						}

						hasTeeTimeLock1AOnRedis := false

						// Lấy số slot đã book
						teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(bookingDateF, body.CourseCode, hourStr, teeType)
						rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
						rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
						//

						// Get số slot tee time còn trống
						teeTimeSlotEmptyRedisKey := getKeyTeeTimeSlotRedis(bookingDateF, body.CourseCode, hourStr, teeType)
						slotStr, _ := datasources.GetCache(teeTimeSlotEmptyRedisKey)
						slotLock, _ := strconv.Atoi(slotStr)
						slotEmpty := slotLock + len(rowIndexsRedis)

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
								TeeType:      teeType,
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
		}

	}

	//List TeeType Info response
	listTeeTypeRes := []response.TeeTypeOTARes{}
	for _, v := range listTeeType {
		teeTypeRes := response.TeeTypeOTARes{
			TeeType:   v.TeeType,
			Name:      v.Name,
			ImageLink: v.ImageLink,
			Note:      v.Note,
		}
		listTeeTypeRes = append(listTeeTypeRes, teeTypeRes)
	}

	responseOTA.Data = teeTimeList
	responseOTA.TeeTypeInfo = listTeeTypeRes
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
			CreatedAt: utils.GetTimeNow().Unix(),
		},
	}

	if body.Tee == "1" {
		teeTimeSetting.TeeType = "1A"
	}

	if body.Tee == "10" {
		teeTimeSetting.TeeType = "1B"
	}

	// Lấy số slot đã book
	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(teeTimeSetting.DateTime, teeTimeSetting.CourseUid, teeTimeSetting.TeeTime, teeTimeSetting.TeeType)
	rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
	//

	teeTimeSlotEmptyRedisKey := getKeyTeeTimeSlotRedis(teeTimeSetting.DateTime, teeTimeSetting.CourseUid, teeTimeSetting.TeeTime, teeTimeSetting.TeeType)
	slotStr, _ := datasources.GetCache(teeTimeSlotEmptyRedisKey)
	slotLockOTA, _ := strconv.Atoi(slotStr)
	slotBook := slotLockOTA + len(rowIndexsRedis)
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
		Type:           constants.LOCK_OTA,
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
			if errSlotEmpty := datasources.SetCache(teeTimeSlotEmptyRedisKey, body.NumBook+slotLockOTA, 0); errSlotEmpty != nil {
				log.Println("LockTeeTimeOta ", errSlotEmpty.Error())
			} else {
				// Bắn socket để client update ui
				go func() {
					cNotification := CNotification{}
					cNotification.PushNotificationLockTee(constants.NOTIFICATION_LOCK_TEE)
				}()
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

	errorCommon := func() {
		responseOTA.Result.Status = http.StatusInternalServerError
		responseOTA.Result.Infor = "Có lỗi xảy ra!"
		c.JSON(http.StatusInternalServerError, responseOTA)
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
	lockTeeRedisRaw, errTimelock := datasources.GetCache(teeTimeRedisKey)
	lockTeeTime := models.LockTeeTimeWithSlot{}

	if errTimelock != nil {
		errorCommon()
		return
	} else {
		if err := json.Unmarshal([]byte(lockTeeRedisRaw), &lockTeeTime); err != nil {
			errorCommon()
			return
		}
	}

	teeTimeSlotEmptyRedisKey := getKeyTeeTimeSlotRedis(dateFormat, body.CourseCode, body.TeeOffStr, "1A")
	slotStr, _ := datasources.GetCache(teeTimeSlotEmptyRedisKey)
	slotLock, _ := strconv.Atoi(slotStr)
	slotRemain := slotLock - body.NumBook

	if slotRemain > 0 {
		if err := datasources.SetCache(teeTimeSlotEmptyRedisKey, slotRemain, 0); err != nil {
			log.Print("updateSlotTeeTime", err)
		}

		lockTeeTime.Slot = slotRemain
		valueParse, _ := lockTeeTime.Value()

		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 0); err != nil {
			log.Print("updateSlotTeeTime", err)
		}
	} else {

		errTeeTime := datasources.DelCacheByKey(teeTimeRedisKey)
		log.Print("runCheckLockTeeTime", errTeeTime)

		err := datasources.DelCacheByKey(teeTimeSlotEmptyRedisKey)
		log.Print("runCheckLockTeeTime", err)
	}
	// Bắn socket để client update ui
	go func() {
		cNotification := CNotification{}
		cNotification.PushNotificationLockTee(constants.NOTIFICATION_UNLOCK_TEE)
	}()
	okResponse(c, responseOTA)
}
