package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"time"

	model_booking "start/models/booking"
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
	//Lấy danh sách booking trong ngày
	now := time.Now().Format("02/01/2006")

	bookingRequest := model_booking.Booking{}
	listBooking, err := bookingRequest.FindAllBookingCheckIn(now)

	if err != nil {
		log.Println("runCreateCaddieFeeOnDay err", err.Error())
		return
	}

	for _, v := range listBooking {
		if v.CaddieId > 0 {
			// get caddie fee group setting today
			date := utils.GetTimeStampFromLocationTime("", constants.DATE_FORMAT_1, now)

			caddieFeeGroupSetting := models.CaddieFeeSettingGroup{}
			caddieFeeGroupSetting.PartnerUid = v.PartnerUid
			caddieFeeGroupSetting.CourseUid = v.CourseUid

			err = caddieFeeGroupSetting.FindFirstByDate(date)
			if err != nil {
				log.Println("get caddie fee setting group err", err.Error())
				return
			}

			// get list caddie setiing by group
			caddieFeeSetting := models.CaddieFeeSetting{}
			caddieFeeSetting.PartnerUid = v.PartnerUid
			caddieFeeSetting.CourseUid = v.CourseUid
			caddieFeeSetting.GroupId = caddieFeeGroupSetting.Id

			listCFSeting, err := caddieFeeSetting.FindAll()
			if err != nil {
				log.Println("get  list caddie fee setting err", err.Error())
				return
			}

			// Check caddie is exist
			caddieFee := models.CaddieFee{}
			caddieFee.CaddieId = v.CaddieId
			caddieFee.BookingDate = now
			err = caddieFee.FindFirst()

			if err != nil {
				// create caddie fee
				for _, cfs := range listCFSeting {
					if cfs.Hole == v.Hole {
						caddieFee.Amount = cfs.Fee
					}
				}

				caddieFee.PartnerUid = v.PartnerUid
				caddieFee.CourseUid = v.CourseUid
				caddieFee.CaddieCode = v.CaddieInfo.Code
				caddieFee.CaddieName = v.CaddieInfo.Name
				caddieFee.Hole = v.Hole
				caddieFee.Round = 1

				err = caddieFee.Create()
				if err != nil {
					log.Println("Create report caddie err", err.Error())
					return
				}
			} else {
				// update caddie fee
				for _, cfs := range listCFSeting {
					if cfs.Hole == v.Hole {
						caddieFee.Amount += cfs.Fee
					}
				}

				caddieFee.Hole += v.Hole
				caddieFee.Round += 1

				err = caddieFee.Update()
				if err != nil {
					log.Println("Create report caddie err", err.Error())
					return
				}
			}
		}
	}
}
