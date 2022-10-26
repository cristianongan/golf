package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	model_gostarter "start/models/go-starter"
	"start/utils"
	"time"

	"gorm.io/datatypes"
)

func runReportCaddieFeeToDayJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerReportCaddieFeeToDay(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runReportCaddieFeeToDay()
}

// Báo cáo số fee của caddie trong ngày,
func runReportCaddieFeeToDay() {
	//Lấy danh sách caddie in out note trong ngày
	now := time.Now().Format("02/01/2006")
	nowUnix := time.Now().Unix()

	log.Println("runReportCaddieFeeToDay", time.Now().UnixNano())
	db := datasources.GetDatabase()

	caddies := models.Caddie{}
	listcaddies, err := caddies.FindAllCaddieContract(db)

	if err != nil {
		log.Println("runCreateCaddieFeeOnDay err", err.Error())
		return
	}

	for _, v := range listcaddies {
		// get group caddie
		groupCaddie := models.CaddieGroup{}

		groupCaddie.Id = v.GroupId
		errGC := groupCaddie.FindFirst(db)
		if errGC != nil {
			log.Println("Find frist group caddie", errGC.Error())
			return
		}

		// get Date
		dateConvert, _ := time.Parse(constants.DATE_FORMAT_1, now)
		applyDate := datatypes.Date(dateConvert)

		// get caddie work sechedule
		caddieWorkingSchedule := models.CaddieWorkingSchedule{
			CaddieGroupCode: groupCaddie.Code,
			ApplyDate:       &(applyDate),
		}

		errCWS := caddieWorkingSchedule.FindFirst(db)
		if errCWS != nil {
			log.Println("Find frist caddie working schedule", errCWS.Error())
			return
		}

		// get caddie fee group setting today
		date := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, now)

		caddieFeeGroupSetting := models.CaddieFeeSettingGroup{}
		caddieFeeGroupSetting.PartnerUid = v.PartnerUid
		caddieFeeGroupSetting.CourseUid = v.CourseUid

		err = caddieFeeGroupSetting.FindFirstByDate(db, date)
		if err != nil {
			log.Println("get caddie fee setting group err", err.Error())
			return
		}

		// get list caddie setiing by group
		caddieFeeSetting := models.CaddieFeeSetting{}
		caddieFeeSetting.PartnerUid = v.PartnerUid
		caddieFeeSetting.CourseUid = v.CourseUid
		caddieFeeSetting.GroupId = caddieFeeGroupSetting.Id

		listCFSeting, err := caddieFeeSetting.FindAll(db)
		if err != nil {
			log.Println("get  list caddie fee setting err", err.Error())
			return
		}

		caddieIONRequest := model_gostarter.CaddieBuggyInOut{}
		caddieIONRequest.CaddieId = v.Id
		listCaddieION, err := caddieIONRequest.FindAllCaddieInOutNotes(db, now)

		if err != nil {
			log.Println("runCreateCaddieFeeOnDay err", err.Error())
			return
		}

		// Create caddie fee
		caddieFee := models.CaddieFee{}
		caddieFee.PartnerUid = v.PartnerUid
		caddieFee.CourseUid = v.CourseUid
		caddieFee.CaddieId = v.Id
		caddieFee.BookingDate = now
		caddieFee.CaddieCode = v.Code
		caddieFee.CaddieName = v.Name
		caddieFee.IsDayOff = false

		if len(listCaddieION) > 0 {
			for _, item := range listCaddieION {
				// create caddie fee
				for _, cfs := range listCFSeting {
					if cfs.Hole >= item.Hole && item.Hole > 0 {
						caddieFee.Amount += cfs.Fee
						caddieFee.Round += 1
						break
					}
				}
			}
		}

		idDayOff := true
		if caddieWorkingSchedule.IsDayOff == &idDayOff {
			caddieFee.IsDayOff = true
			caddieFee.Note = "Tăng cường"
		}

		//Kiểm tra nhày nghỉ của caddie
		caddieVC := models.CaddieVacationCalendar{}
		caddieVC.PartnerUid = v.PartnerUid
		caddieVC.CourseUid = v.CourseUid
		caddieVC.CaddieId = v.Id
		caddieVC.DateFrom = nowUnix

		listItem, errCVC := caddieVC.FindAll(db)

		if errCVC != nil {
			log.Println("Find caddie vacation calendar err", err.Error())
			return
		}

		if len(listItem) > 0 {
			caddieFee.Note = "Nghỉ phép"
		}

		err = caddieFee.Create(db)
		if err != nil {
			log.Println("Create report caddie err", err.Error())
			return
		}
	}
}
