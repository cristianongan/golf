package cron

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
)

func runResetCaddieStatusJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemResetCaddieStatus(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runResetCaddieStatus()
}

// Reset số guest của member trong ngày
func runResetCaddieStatus() {
	localTimeTomorow, _ := utils.GetLocalTimeFromTimeStamp(constants.LOCATION_DEFAULT, constants.DATE_FORMAT_1, utils.GetTimeNow().AddDate(0, 0, 1).Unix())

	caddie := models.CaddieList{}
	dbCaddie := datasources.GetDatabase()
	listCaddie, _, _ := caddie.FindAllCaddieReadyOnDayList(dbCaddie) // Lấy ra caddie trong ngày làm việc
	/*
		Reset het trang thai cua nhung thang do
	*/
	for _, v := range listCaddie {
		checkSlot := false
		//
		caddieWS := models.CaddieWorkingSlot{}
		caddieWS.PartnerUid = v.PartnerUid
		caddieWS.CourseUid = v.CourseUid
		caddieWS.ApplyDate = localTimeTomorow

		_ = caddieWS.FindFirst(dbCaddie)

		if len(caddieWS.CaddieSlot) > 0 {
			checkSlot = utils.Contains(caddieWS.CaddieSlot, v.Code)
		}

		caddieWC := models.CaddieWorkingCalendar{}
		caddieWC.PartnerUid = v.PartnerUid
		caddieWC.CourseUid = v.CourseUid
		caddieWC.ApplyDate = localTimeTomorow
		caddieWC.CaddieCode = v.Code

		_ = caddieWC.FindFirst(dbCaddie)

		if caddieWC.NumberOrder > 0 {
			checkSlot = true
		}

		v.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
		v.CurrentRound = 0

		if !checkSlot {
			v.IsWorking = 0
		}
		v.Update(dbCaddie)
		updateCaddieOutSlot(v.PartnerUid, v.CourseUid, []string{v.Code})
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
