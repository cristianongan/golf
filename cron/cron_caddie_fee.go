package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	model_gostarter "start/models/go-starter"
	"start/utils"
	"time"
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

	log.Println("runReportCaddieFeeToDay", time.Now().UnixNano())
	db := datasources.GetDatabase()

	caddieIONRequest := model_gostarter.CaddieInOutNote{}
	listCaddieION, err := caddieIONRequest.FindAllCaddieInOutNotes(db)

	if err != nil {
		log.Println("runCreateCaddieFeeOnDay err", err.Error())
		return
	}

	for _, v := range listCaddieION {
		if v.CaddieId > 0 {
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

			// Check caddie is exist
			caddieFee := models.CaddieFee{}
			caddieFee.CaddieId = v.CaddieId
			caddieFee.BookingDate = now
			err = caddieFee.FindFirst(db)

			if err != nil {
				// find caddie name
				caddie := models.Caddie{}
				caddie.Id = v.CaddieId
				err = caddie.FindFirst(db)
				if err != nil {
					log.Println("find first caddie err", err.Error())
					return
				}

				// create caddie fee
				for _, cfs := range listCFSeting {
					if cfs.Hole >= v.Hole && v.Hole > 0 {
						caddieFee.Amount += cfs.Fee
						break
					}
				}

				caddieFee.PartnerUid = v.PartnerUid
				caddieFee.CourseUid = v.CourseUid
				caddieFee.CaddieCode = v.CaddieCode
				caddieFee.CaddieName = caddie.Name
				caddieFee.Hole = v.Hole
				caddieFee.Round = 1

				err = caddieFee.Create(db)
				if err != nil {
					log.Println("Create report caddie err", err.Error())
					return
				}
			} else {
				// update caddie fee
				for _, cfs := range listCFSeting {
					if cfs.Hole >= v.Hole && v.Hole > 0 {
						caddieFee.Amount += cfs.Fee
						break
					}
				}

				caddieFee.Hole += v.Hole
				if caddieFee.Hole > 0 {
					caddieFee.Round += 1
				}

				err = caddieFee.Update(db)
				if err != nil {
					log.Println("Update report caddie err", err.Error())
					return
				}
			}
		}
	}
}
