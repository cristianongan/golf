package cron

import (
	"log"
	"start/config"

	"github.com/robfig/cron/v3"
)

func CronStart() {
	c := cron.New()

	c.AddFunc("@every 5s", runCheckLockTeeTime)
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 05 23 * * *", runReportCaddieFeeToDay)            // Chạy lúc 23h00 hàng ngày
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 19 * * *", runCreateCaddieWorkingSlotJob)      // Chạy lúc 18h45 hàng ngày
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 45 18 * * *", runResetCaddieStatusJob)            // Chạy lúc 18h30 tối hàng ngày để reset caddie status
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 10 00 * * *", runResetDataMemberCardJob)          // Chạy lúc 00h10 sáng hàng ngày để reset data trong ngày của member card
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 00 * * *", runReportInventoryStatisticItemJob) // Chạy lúc 0h sáng hàng ngày để thống kê sản phẩm trong kiosk inventory
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 * * * *", runEnableUserJob)                    // Chạy hàng giờ để enable user bị khóa
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 23 * * *", runBookingLogutJob)                 // Chạy lúc 23h30 tối hàng ngày để logout các booking chưa checkout
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 00 23 * * *", runResetBuggyStatusJob)             // Chạy lúc 23h30 tối hàng ngày để reset status buggy
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 30 23 * * *", runReportDailyRevenueJob)           // Chạy hàng ngày report DT
	// Add tiếp các cron khác dưới đây

	// Check config có chạy Cron hay không
	if config.GetCronIsRunning() {
		log.Println("==== CronStart =====")
		c.Start()
	}
}
