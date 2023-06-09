package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	model_gostarter "start/models/go-starter"
	"start/utils"
	"time"

	"github.com/bsm/redislock"
	"gorm.io/datatypes"
)

func runReportCaddieFeeToDayJob() {
	// Để xử lý cho chạy nhiều instance Server
	// isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerReportCaddieFeeToDay(), 60)
	// // Ko lấy được lock, return luôn
	// if !isObtain {
	// 	return
	// }

	redisKey := datasources.GetRedisKeyLockerReportCaddieFeeToDay()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 60*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[CRON] runReportCaddieFeeToDayJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

	// Logic chạy cron bên dưới
	runReportCaddieFeeToDay()
}

// Báo cáo số fee của caddie trong ngày,
func runReportCaddieFeeToDay() {
	//Lấy danh sách caddie in out note trong ngày
	now := utils.GetTimeNow().Format("02/01/2006")
	day := utils.GetTimeNow().Weekday()
	nowUnix := utils.GetTimeNow().Unix()

	log.Println("runReportCaddieFeeToDay", utils.GetTimeNow().UnixNano())
	db := datasources.GetDatabase()

	caddies := models.Caddie{}
	listcaddies, err := caddies.FindAllCaddieContract(db)

	if err != nil {
		log.Println("runCreateCaddieFeeOnDay err", err.Error())
	}

	for _, v := range listcaddies {
		// get group caddie
		groupCaddie := models.CaddieGroup{}

		groupCaddie.Id = v.GroupId
		errGC := groupCaddie.FindFirst(db)
		if errGC != nil {
			log.Println("Find frist group caddie", errGC.Error())
		}

		// get Date
		dateConvert, _ := time.Parse(constants.DATE_FORMAT_1, now)
		applyDate := datatypes.Date(dateConvert)
		idDayOff := true

		// get caddie work sechedule
		caddieWorkingSchedule := models.CaddieWorkingSchedule{
			CaddieGroupCode: groupCaddie.Code,
			ApplyDate:       &(applyDate),
			IsDayOff:        &idDayOff,
		}

		errCWS := caddieWorkingSchedule.FindFirst(db)
		if errCWS != nil {
			log.Println("Find frist caddie working schedule", errCWS.Error())
		}

		// get caddie fee group setting today
		date := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, now)

		caddieFeeGroupSetting := models.CaddieFeeSettingGroup{}
		caddieFeeGroupSetting.PartnerUid = v.PartnerUid
		caddieFeeGroupSetting.CourseUid = v.CourseUid

		err = caddieFeeGroupSetting.FindFirstByDate(db, date)
		if err != nil {
			log.Println("get caddie fee setting group err", err.Error())
		}

		// get list caddie setiing by group
		caddieFeeSetting := models.CaddieFeeSetting{}
		caddieFeeSetting.PartnerUid = v.PartnerUid
		caddieFeeSetting.CourseUid = v.CourseUid
		caddieFeeSetting.GroupId = caddieFeeGroupSetting.Id

		listCFSeting, err := caddieFeeSetting.FindAll(db)
		if err != nil {
			log.Println("get  list caddie fee setting err", err.Error())
		}

		caddieIONRequest := model_gostarter.CaddieBuggyInOut{}
		caddieIONRequest.CaddieId = v.Id
		listCaddieION, err := caddieIONRequest.FindAllCaddieInOutNotes(db, now)

		if err != nil {
			log.Println("runCreateCaddieFeeOnDay err", err.Error())
		}

		// Create caddie fee
		caddieFee := models.CaddieFee{}
		caddieFee.PartnerUid = v.PartnerUid
		caddieFee.CourseUid = v.CourseUid
		caddieFee.CaddieId = v.Id
		caddieFee.BookingDate = now
		caddieFee.CaddieCode = v.Code
		caddieFee.CaddieName = v.Name

		if len(listCaddieION) > 0 {
			for _, item := range listCaddieION {
				// create caddie fee
				for _, cfs := range listCFSeting {
					if cfs.Hole >= item.Hole && item.Hole > 0 {
						caddieFee.Hole += item.Hole
						caddieFee.Amount += cfs.Fee
						caddieFee.Round += 1

						break
					}
				}
			}
		}

		if caddieWorkingSchedule.Id > 0 &&
			(int(day) != 0 && int(day) != 6) {
			caddieFee.IsDayOff = caddieWorkingSchedule.IsDayOff
			if len(listCaddieION) > 0 {
				caddieFee.Note = "Tăng cường"
			}
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
		}

		if len(listItem) > 0 {
			caddieFee.Note = "Nghỉ phép"
		}

		if !caddieFee.IsDuplicated(db) {
			err = caddieFee.Create(db)
			if err != nil {
				log.Println("Create report caddie err", err.Error())
			}
		}
	}
}
