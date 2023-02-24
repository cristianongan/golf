package cron

import (
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
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
	localTime, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, utils.GetTimeNow().Unix())
	bookingList := model_booking.BookingList{
		BookingDate: localTime,
		IsCheckIn:   "1",
	}

	dbBooking1, _, _ = bookingList.FindAllBookingList(dbBooking1)

	list := []model_booking.Booking{}
	dbBooking1.Find(&list)

	// dbBooking2 := datasources.GetDatabase()
	// for _, booking := range list {
	// 	booking.BagStatus = constants.BAG_STATUS_CHECK_OUT
	// 	booking.CheckOutTime = utils.GetTimeNow().Unix()
	// 	if err := booking.Update(dbBooking2); err != nil {
	// 		log.Print(err.Error())
	// 	}
	// }

	caddie := models.CaddieList{}
	dbCaddie := datasources.GetDatabase()
	listCaddie, _, _ := caddie.FindAllCaddieReadyOnDayList(dbCaddie) // Lấy ra caddie trong ngày làm việc
	/*
		Reset het trang thai cua nhung thang do
	*/
	for _, v := range listCaddie {
		v.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
		v.CurrentRound = 0
		v.IsWorking = 0
		v.Update(dbCaddie)
		updateCaddieOutSlot(v.PartnerUid, v.CourseUid, []string{v.Code})
	}

	buggy := models.Buggy{}
	dbBuggy := datasources.GetDatabase()
	listBuggy, _, _ := buggy.FindListBuggyNotReady(dbBuggy)
	for _, v := range listBuggy {
		v.BuggyStatus = constants.BUGGY_CURRENT_STATUS_ACTIVE
		v.Update(dbBuggy)
	}
}

func updateCaddieOutSlot(partnerUid, courseUid string, caddies []string) error {
	var caddieSlotNew []string
	var caddieSlotExist []string
	// Format date
	dateNow, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	caddieWS := models.CaddieWorkingSlot{}
	caddieWS.PartnerUid = partnerUid
	caddieWS.CourseUid = courseUid
	caddieWS.ApplyDate = dateNow

	db := datasources.GetDatabaseWithPartner(partnerUid)

	err := caddieWS.FindFirst(db)
	if err != nil {
		return err
	}

	if len(caddieWS.CaddieSlot) > 0 {
		caddieSlotNew = append(caddieSlotNew, caddieWS.CaddieSlot...)
		for _, item := range caddies {
			index := utils.StringInList(item, caddieSlotNew)
			if index != -1 {
				caddieSlotNew = utils.Remove(caddieSlotNew, index)
				caddieSlotExist = append(caddieSlotExist, item)
			}
		}
	}

	caddieWS.CaddieSlot = append(caddieSlotNew, caddieSlotExist...)
	err = caddieWS.Update(db)
	if err != nil {
		return err
	}

	return nil
}
