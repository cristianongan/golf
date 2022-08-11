package cron

import (
	"log"
	"start/datasources"
	"start/models"
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
	bookingRequest := model_booking.Booking{}
	listBooking, err := bookingRequest.FindAll(time.Now().Format("02/06/2006"))

	if err != nil {
		log.Println("runCreateCaddieFeeOnDay err", err.Error())
		return
	}

	log.Println("==== log caddie =====", listBooking)

	for _, v := range listBooking {
		if v.CaddieId > 0 {
			// Check caddie is
			caddie := models.CaddieFee{}
			caddie.CaddieId = v.CaddieId
			err = caddie.FindFirst()

			if err != nil {
				log.Println("runCreateCaddieFeeOnDay err", err.Error())
				return
			}

			log.Println("==== log caddie =====", caddie)
		}
	}
}
