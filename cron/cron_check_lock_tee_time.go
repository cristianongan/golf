package cron

import (
	"encoding/json"
	"log"
	"start/config"
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
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
			teeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:"
			teeTimeRedisKey += teeTime.DateTime + "_" + teeTime.CourseUid + "_" + teeTime.TeeTime + "_" + teeTime.TeeType

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

			prefixRedisKey := config.GetEnvironmentName() + "tee_time_lock:" + teeTime.DateTime + "_" + teeTime.CourseUid + "_"
			prefixRedisKey += teeTime.TeeTime + "_" + teeType + courseType

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

			for _, item := range listTeeTimeLockRedis {
				total -= int64(item.Slot)
			}

			err := datasources.DelCacheByKey(teeTimeRedisKey)
			log.Print("runCheckLockTeeTime", err)

			slotTeeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty" + "_" + teeTime.CourseUid + "_" + teeTime.DateTime + "_" + teeTime.TeeType + "_" + teeTime.TeeTime
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
