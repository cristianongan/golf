package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strconv"
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
