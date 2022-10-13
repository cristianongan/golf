package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"time"
)

func runBookingLogutJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemLogout(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runBookingLogout()
}

func runBookingLogout() {
	db := datasources.GetDatabase()
	localTime, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, time.Now().Unix())
	bookingList := model_booking.BookingList{
		BookingDate: localTime,
		IsCheckIn:   "1",
	}

	db, _, _ = bookingList.FindAllBookingList(db)

	list := []model_booking.Booking{}
	db.Find(&list)

	for _, booking := range list {
		booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
		booking.CheckOutTime = time.Now().Unix()

		if err := booking.Update(db); err != nil {
			log.Print("cron update booking check error!")
		}
	}

	caddie := models.Caddie{}
	listCaddie, _, _ := caddie.FindListCaddieNotReady(db)
	for _, v := range listCaddie {
		v.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
		v.Update(db)
	}

	buggy := models.Buggy{}
	listBuggy, _, _ := buggy.FindListBuggyNotReady(db)
	for _, v := range listBuggy {
		v.BuggyStatus = constants.BUGGY_CURRENT_STATUS_ACTIVE
		v.Update(db)
	}
}
