package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"strconv"
	"strings"
	"time"
)

/*
Param: teeType loại tee time, numBook số lượng book, bookingDate ngày book, teeTime thời gian tee, courseUid mã sân, note ghi chú lock
*/
func lockTee(teeType string, numBook int, bookingDate string, teeTime string, courseUid string, note string) error {
	teeTimeSetting := models.LockTeeTime{
		DateTime:       bookingDate,
		CourseUid:      courseUid,
		TeeTime:        teeTime,
		CurrentTeeTime: teeTime,
		TeeType:        teeType,
		ModelId: models.ModelId{
			CreatedAt: utils.GetTimeNow().Unix(),
		},
	}

	// Lấy số slot đã book
	teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(teeTimeSetting.DateTime, teeTimeSetting.CourseUid, teeTimeSetting.TeeTime, teeTimeSetting.TeeType)
	rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
	rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)

	teeTimeSlotEmptyRedisKey := getKeyTeeTimeSlotRedis(teeTimeSetting.DateTime, teeTimeSetting.CourseUid, teeTimeSetting.TeeTime, teeTimeSetting.TeeType)
	slotStr, _ := datasources.GetCache(teeTimeSlotEmptyRedisKey)
	slotLockOTA, _ := strconv.Atoi(slotStr)
	slotBook := slotLockOTA + len(rowIndexsRedis)
	slotEmpty := constants.SLOT_TEE_TIME - slotBook

	if numBook > slotEmpty {
		return errors.New("Slot lock invalid!")
	}

	// Create redis key tee time lock
	teeTimeRedisKey := getKeyTeeTimeLockRedis(bookingDate, courseUid, teeTime, teeType)

	key := datasources.GetRedisKeyTeeTimeLock(teeTimeRedisKey)
	_, errRedis := datasources.GetCache(key)

	teeTimeRedis := models.LockTeeTimeWithSlot{
		DateTime:       teeTimeSetting.DateTime,
		CourseUid:      teeTimeSetting.CourseUid,
		TeeTime:        teeTimeSetting.TeeTime,
		CurrentTeeTime: teeTimeSetting.TeeTime,
		TeeType:        teeTimeSetting.TeeType,
		TeeTimeStatus:  constants.TEE_TIME_LOCKED,
		Slot:           numBook,
		Type:           constants.LOCK_OTA,
		Note:           note,
		ModelId: models.ModelId{
			CreatedAt: teeTimeSetting.CreatedAt,
		},
	}

	if errRedis != nil {
		valueParse, _ := teeTimeRedis.Value()
		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 0); err != nil {
			return errors.New(err.Error())
		} else {
			if errSlotEmpty := datasources.SetCache(teeTimeSlotEmptyRedisKey, numBook+slotLockOTA, 0); errSlotEmpty != nil {
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

	return nil
}

func searchTeeTimeList(courseUid string, dateStr string, tokenStr string, agencyId string, hole int, teeType string, isMainCourse bool) ([]response.TeeTimeOTA, []models.TeeTypeInfo, error) {
	date, _ := time.Parse("2006-01-02", dateStr)
	bookingDateF := date.Format("02/01/2006")

	// Find Course
	course := models.Course{}
	course.Uid = courseUid
	if errCourse := course.FindFirstHaveKey(); errCourse != nil {
		return []response.TeeTimeOTA{}, []models.TeeTypeInfo{}, errCourse
	}

	checkToken := course.ApiKey + courseUid + dateStr
	token := utils.GetSHA256Hash(checkToken)

	if strings.ToUpper(token) != tokenStr {
		return []response.TeeTimeOTA{}, []models.TeeTypeInfo{}, errors.New("token is invalid")
	}

	db := datasources.GetDatabase()

	// responseOTA.GolfPriceRate = course.RateGolfFee

	// Lấy Fee
	agency := models.Agency{}
	agency.AgencyId = agencyId
	agency.CourseUid = courseUid
	errFindAgency := agency.FindFirst(db)
	if errFindAgency != nil || agency.Id == 0 {
		// responseOTA.Result.Status = 500
		// responseOTA.Result.Infor = errFindAgency.Error()
		// okResponse(c, responseOTA)
		return []response.TeeTimeOTA{}, []models.TeeTypeInfo{}, errFindAgency
	}
	agencySpecialPriceR := models.AgencySpecialPrice{
		AgencyId:  agency.Id,
		CourseUid: courseUid,
	}

	GreenFee := int64(0)
	CaddieFee := int64(0)
	BuggyFee := int64(0)

	timeDate, _ := time.Parse(constants.DATE_FORMAT, dateStr)
	agencySpecialPrice, errFSP := agencySpecialPriceR.FindOtherPriceOnDate(db, timeDate)
	if errFSP == nil && agencySpecialPrice.Id > 0 {
		GreenFee = agencySpecialPrice.GreenFee
		CaddieFee = agencySpecialPrice.CaddieFee
		// BuggyFee = agencySpecialPrice.BuggyFee
	} else {
		golfFee := models.GolfFee{
			GuestStyle: agency.GuestStyle,
			CourseUid:  courseUid,
		}

		fee, _ := golfFee.GetGuestStyleOnTime(db, timeDate)

		// Lấy giá hole
		GreenFee = utils.GetFeeFromListFee(fee.GreenFee, hole)
		CaddieFee = utils.GetFeeFromListFee(fee.CaddieFee, hole)
		// BuggyFee = utils.GetFeeFromListFee(fee.BuggyFee, body.Hole)
	}

	BuggyFee = utils.GetFeeFromListFee(getBuggyFee(agency.GuestStyle, bookingDateF, course.PartnerUid, course.Uid), 18)

	// Get Setting để tạo list tee time
	cBookingSetting := CBookingSetting{}
	form := request.GetListBookingSettingForm{
		CourseUid: courseUid,
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
		// responseOTA.Result.Status = http.StatusInternalServerError
		// responseOTA.Result.Infor = "TeeType Info not yet config"

		// okResponse(c, responseOTA)
		return []response.TeeTimeOTA{}, []models.TeeTypeInfo{}, errLTP

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

	if teeType != "" && validateTeeType(teeType, listTeeType) {
		// get các teetime đang bị khóa ở redis
		listTeeTimeLockRedis := getTeeTimeLockRedis(courseUid, bookingDateF, teeType)

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

					teeOff, _ := time.Parse(constants.DATE_FORMAT_3, dateStr+" "+hourStr)
					teeOffStr := teeOff.Format("2006-01-02T15:04:05")

					teeTime1 := models.LockTeeTime{
						TeeTime:   hourStr,
						DateTime:  bookingDateF,
						CourseUid: courseUid,
						TeeType:   teeType,
					}

					hasTeeTimeLock1AOnRedis := false

					// Lấy số slot đã book
					teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(bookingDateF, courseUid, hourStr, teeType)
					rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
					rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
					//

					// Get số slot tee time còn trống
					teeTimeSlotEmptyRedisKey := getKeyTeeTimeSlotRedis(bookingDateF, courseUid, hourStr, teeType)
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
							DateStr:      dateStr,
							TeeOff:       teeOffStr,
							Tee:          1,
							Part:         int64(partIndex),
							TimeIndex:    int64(index),
							NumBook:      int64(constants.SLOT_TEE_TIME - slotEmpty),
							IsMainCourse: isMainCourse,
							GreenFee:     GreenFee,
							CaddieFee:    CaddieFee,
							BuggyFee:     BuggyFee,
							Holes:        int64(hole),
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
	} else {
		for _, teeTypeInfo := range listTeeType {
			teeType := teeTypeInfo.TeeType
			// get các teetime đang bị khóa ở redis
			listTeeTimeLockRedis := getTeeTimeLockRedis(courseUid, bookingDateF, teeType)

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

						teeOff, _ := time.Parse(constants.DATE_FORMAT_3, dateStr+" "+hourStr)
						teeOffStr := teeOff.Format("2006-01-02T15:04:05")

						teeTime1 := models.LockTeeTime{
							TeeTime:   hourStr,
							DateTime:  bookingDateF,
							CourseUid: courseUid,
							TeeType:   teeType,
						}

						hasTeeTimeLock1AOnRedis := false

						// Lấy số slot đã book
						teeTimeRowIndexRedis := getKeyTeeTimeRowIndex(bookingDateF, courseUid, hourStr, teeType)
						rowIndexsRedisStr, _ := datasources.GetCache(teeTimeRowIndexRedis)
						rowIndexsRedis := utils.ConvertStringToIntArray(rowIndexsRedisStr)
						//

						// Get số slot tee time còn trống
						teeTimeSlotEmptyRedisKey := getKeyTeeTimeSlotRedis(bookingDateF, courseUid, hourStr, teeType)
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
								DateStr:      dateStr,
								TeeOff:       teeOffStr,
								Tee:          1,
								Part:         int64(partIndex),
								TimeIndex:    int64(index),
								NumBook:      int64(constants.SLOT_TEE_TIME - slotEmpty),
								IsMainCourse: isMainCourse,
								GreenFee:     GreenFee,
								CaddieFee:    CaddieFee,
								BuggyFee:     BuggyFee,
								Holes:        int64(hole),
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

	return teeTimeList, listTeeType, nil

}
