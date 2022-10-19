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
	dbBooking1 := datasources.GetDatabase()
	localTime, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, time.Now().Unix())
	bookingList := model_booking.BookingList{
		BookingDate: localTime,
		IsCheckIn:   "1",
	}

	dbBooking1, _, _ = bookingList.FindAllBookingList(dbBooking1)

	list := []model_booking.Booking{}
	dbBooking1.Find(&list)

	dbBooking2 := datasources.GetDatabase()
	for _, booking := range list {
		booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
		booking.CheckOutTime = time.Now().Unix()
		if err := booking.Update(dbBooking2); err != nil {
			log.Print(err.Error())
		}
	}

	caddie := models.Caddie{}
	dbCaddie := datasources.GetDatabase()
	listCaddie, _, _ := caddie.FindListCaddieNotReady(dbCaddie)
	for _, v := range listCaddie {
		v.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
		v.Update(dbCaddie)
	}

	buggy := models.Buggy{}
	dbBuggy := datasources.GetDatabase()
	listBuggy, _, _ := buggy.FindListBuggyNotReady(dbBuggy)
	for _, v := range listBuggy {
		v.BuggyStatus = constants.BUGGY_CURRENT_STATUS_ACTIVE
		v.Update(dbBuggy)
	}
}
