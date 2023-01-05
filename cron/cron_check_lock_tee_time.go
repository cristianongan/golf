package cron

import (
	"encoding/json"
	"log"
	"start/config"
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/socket"
	"start/utils"
	"time"
)

func runCheckLockTeeTime() {
	today, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())
	prefixRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:" + today
	listKey, errRedis := datasources.GetAllKeysWith(prefixRedisKey)
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

	constantTime := 5 * 60
	for _, teeTime := range listTeeTimeLockRedis {
		diff := time.Now().Unix() - teeTime.CreatedAt
		if diff >= int64(constantTime) && teeTime.Type == constants.BOOKING_OTA {

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
				pushNotificationCreateBookingOTA(constants.NOTIFICATION_BOOKING_OTA)
			}()

			slotTeeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty:" + teeTime.CourseUid + "_" + teeTime.DateTime + "_" + teeTime.TeeType + "_" + teeTime.TeeTime
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
}

func getKeyTeeTimeLockRedis(bookingDate, courseUid, teeTime, teeType string) string {
	teeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:" + bookingDate + "_" + courseUid + "_"
	teeTimeRedisKey += teeType + "_" + teeTime

	return teeTimeRedisKey
}

func pushNotificationCreateBookingOTA(title string) {
	notiData := map[string]interface{}{
		"type":  constants.NOTIFICATION_BOOKING_OTA,
		"title": title,
	}

	newFsConfigBytes, _ := json.Marshal(notiData)
	socket.HubBroadcastSocket.Broadcast <- newFsConfigBytes
}
