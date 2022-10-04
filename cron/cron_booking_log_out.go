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

		// udp trạng thái caddie
		dbCaddie := datasources.GetDatabase()
		caddie := models.Caddie{}
		caddie.Id = booking.CaddieId
		if err := caddie.FindFirst(dbCaddie); err == nil {
			if caddie.CurrentStatus != constants.CADDIE_CURRENT_STATUS_READY {
				caddie.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
				if errUdp := caddie.Update(dbCaddie); errUdp != nil {
					log.Println("udpBuggyOut err", err.Error())
				}
			}
		}
		// udp trạng thái buggy
		dbBuggy := datasources.GetDatabase()
		buggy := models.Buggy{}
		buggy.Id = booking.BuggyId

		if err := buggy.FindFirst(dbBuggy); err == nil {
			if buggy.BuggyStatus != constants.BUGGY_CURRENT_STATUS_ACTIVE {
				buggy.BuggyStatus = constants.BUGGY_CURRENT_STATUS_ACTIVE
				if errUdp := buggy.Update(dbBuggy); errUdp != nil {
					log.Println("udpBuggyOut err", err.Error())
				}
			}
		}
	}
}
