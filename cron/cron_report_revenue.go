package cron

import (
	"log"
	"start/datasources"
	model_booking "start/models/booking"
	model_report "start/models/report"
	"start/utils"
)

func runReportDailyRevenueJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerResetDataMemberCard(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runReportDailyRevenue()
}

// Reset số guest của member trong ngày
func runReportDailyRevenue() {
	db := datasources.GetDatabaseWithPartner("CHI-LINH")

	toDayDate, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	bookings := model_booking.BookingList{
		BookingDate: toDayDate,
	}

	db, _, err := bookings.FindAllBookingList(db)
	db = db.Where("check_in_time > 0")
	db = db.Where("bag_status <> 'CANCEL'")
	db = db.Where("init_type <> 'ROUND'")
	db = db.Where("init_type <> 'MOVEFLGIHT'")

	if err != nil {
		log.Println(err.Error())
	}

	var list []model_booking.Booking
	db.Find(&list)

	reportR := model_report.ReportRevenueDetail{
		PartnerUid:  "CHI-LINH",
		CourseUid:   "CHI-LINH-01",
		BookingDate: toDayDate,
	}

	reportR.DeleteByBookingDate()

	for _, booking := range list {
		updatePriceForRevenue(booking, "")
	}
}
