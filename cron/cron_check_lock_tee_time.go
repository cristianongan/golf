package cron

import (
	"encoding/json"
	"log"
	"start/callservices"
	"start/config"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
)

func runCheckLockTeeTime() {
	prefixRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:"
	listKey, errRedis := datasources.GetAllKeysWith(prefixRedisKey)
	listTeeTimeLockRedis := []models.LockTeeTimeWithSlot{}
	if errRedis == nil && len(listKey) > 0 {
		strData, errGet := datasources.GetCaches(listKey...)
		if errGet != nil {
			log.Println("checkBookingOTA-error", errGet.Error())
		} else {
			for _, data := range strData {
				if data != nil {
					byteData := []byte(data.(string))
					teeTime := models.LockTeeTimeWithSlot{}
					err2 := json.Unmarshal(byteData, &teeTime)
					if err2 == nil {
						listTeeTimeLockRedis = append(listTeeTimeLockRedis, teeTime)
					}
				}
			}
		}
	}

	constantTime := 5 * 60
	for _, teeTime := range listTeeTimeLockRedis {
		diff := utils.GetTimeNow().Local().Unix() - teeTime.CreatedAt
		if diff >= int64(constantTime) && teeTime.Type == constants.LOCK_OTA {

			teeTimeRedisKey := getKeyTeeTimeLockRedis(teeTime.DateTime, teeTime.CourseUid, teeTime.TeeTime, "1A")
			teeType := teeTime.TeeType[0:1]
			courseType := teeTime.TeeType[len(teeTime.TeeType)-1:]

			db := datasources.GetDatabase()
			bookings := model_booking.BookingList{}
			// bookings.PartnerUid = booking.PartnerUid
			bookings.CourseUid = teeTime.CourseUid
			bookings.BookingDate = teeTime.DateTime
			bookings.TeeTime = teeTime.TeeTime
			bookings.TeeType = teeType
			bookings.CourseType = courseType

			_, total, _ := bookings.FindAllBookingNotCancelList(db)

			listKey, errRedis := datasources.GetAllKeysWith(teeTimeRedisKey)
			listTeeTimeLockRedis := []models.LockTeeTimeWithSlot{}
			if errRedis == nil && len(listKey) > 0 {
				strData, errGet := datasources.GetCaches(listKey...)
				if errGet != nil {
					log.Println("checkBookingOTA-error", errGet.Error())
				} else {
					for _, data := range strData {
						if data != nil {
							byteData := []byte(data.(string))
							teeTime := models.LockTeeTimeWithSlot{}
							err2 := json.Unmarshal(byteData, &teeTime)
							if err2 == nil {
								listTeeTimeLockRedis = append(listTeeTimeLockRedis, teeTime)
							}
						}
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
				pushNotificationUnlockTee()
			}()

			slotTeeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty:" + teeTime.DateTime + "_" + teeTime.CourseUid + "_" + teeTime.TeeType + "_" + teeTime.TeeTime
			if total > 0 {
				if err := datasources.SetCache(slotTeeTimeRedisKey, total, 0); err != nil {
					log.Print("updateSlotTeeTime", err)
				}
			} else {
				err := datasources.DelCacheByKey(slotTeeTimeRedisKey)
				log.Print("runCheckLockTeeTime", err)
			}
		}
	}

	// Chú ý: sửa gì ở hàm này thì sửa cả func UnlockTeeTime tên CTeeTimeOTA
}

func getKeyTeeTimeLockRedis(bookingDate, courseUid, teeTime, teeType string) string {
	teeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:" + bookingDate + "_" + courseUid + "_"
	teeTimeRedisKey += teeType + "_" + teeTime

	return teeTimeRedisKey
}

func pushNotificationUnlockTee() {
	notiData := map[string]interface{}{
		"type":  constants.NOTIFICATION_UNLOCK_TEE,
		"title": "",
	}

	// push mess socket
	reqSocket := request.MessSocketBody{
		Data: notiData,
		Room: constants.NOTIFICATION_CHANNEL_BOOKING,
	}

	go callservices.PushMessInSocket(reqSocket)

	// newFsConfigBytes, _ := json.Marshal(notiData)
	// socket_room.Hub.Broadcast <- socket_room.Message{
	// 	Data: newFsConfigBytes,
	// 	Room: constants.NOTIFICATION_CHANNEL_BOOKING,
	// }
}
