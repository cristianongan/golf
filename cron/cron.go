package cron

import (
	"log"
	"start/config"

	"github.com/robfig/cron/v3"
)

func CronStart() {
	c := cron.New()

	// c.AddFunc("@every 1m", runCreateCaddieWorkingSlotJob)
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 05 23 * * *", runReportCaddieFeeToDay) // Chạy lúc 23h00 hàng ngày
	// c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 01 * * *", runCreateCaddieWorkingSlotJob)      // Chạy lúc 1h00 hàng ngày
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 10 00 * * *", runResetDataMemberCardJob)          // Chạy lúc 00h10 sáng hàng ngày để reset data trong ngày của member card
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 00 * * *", runReportInventoryStatisticItemJob) // Chạy lúc 0h sáng hàng ngày để thống kê sản phẩm trong kiosk inventory
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 30 23 * * *", runBookingLogutJob)                 // Chạy lúc 23h30 tối hàng ngày để logout các booking chưa checkout
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 * * * *", runEnableUserJob)                    // Chạy hàng giờ để enable user bị khóa
	// Add tiếp các cron khác dưới đây

	// Check config có chạy Cron hay không
	if config.GetCronIsRunning() {
		log.Println("==== CronStart =====")
		c.Start()
	}
}
