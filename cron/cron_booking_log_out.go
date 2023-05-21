package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	model_booking "start/models/booking"
	"start/utils"
	"time"

	"github.com/bsm/redislock"
)

func runBookingLogutJob() {
	// Để xử lý cho chạy nhiều instance Server
	// isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemLogout(), 60)
	// // Ko lấy được lock, return luôn
	// if !isObtain {
	// 	return
	// }

	redisKey := datasources.GetRedisKeySystemLogout()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 60*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[CRON] runBookingLogutJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

	// Logic chạy cron bên dưới
	runBookingLogout()
}

func runBookingLogout() {
	dbBooking1 := datasources.GetDatabase()
	localTime, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, utils.GetTimeNow().Unix())

	bookingList := model_booking.BookingList{
		BookingDate: localTime,
	}

	dbBooking1, _ = bookingList.FindListBookingNotCheckOut(dbBooking1)

	list := []model_booking.Booking{}
	dbBooking1.Find(&list)

	dbBooking2 := datasources.GetDatabase()
	for _, booking := range list {
		booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
		booking.CheckOutTime = utils.GetTimeNow().Unix()
		if err := booking.Update(dbBooking2); err != nil {
			log.Print(err.Error())
		}
	}
}
