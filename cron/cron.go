package cron

import (
	"log"
	"start/config"

	"github.com/robfig/cron/v3"
)

func CronStart() {
	c := cron.New()

	//c.AddFunc("@every 5s", checkOrderPayment) // 5s chạy 1 lần
	c.AddFunc("CRON_TZ=Asia/Ho_Chi_Minh 10 00 * * *", runResetDataMemberCardJob) // Chạy lúc 00h10 sáng hàng ngày để reset data trong ngày của member card
	// Add tiếp các cron khác dưới đây

	// Check config có chạy Cron hay không
	if config.GetCronIsRunning() {
		log.Println("==== CronStart =====")
		c.Start()
	}
}
