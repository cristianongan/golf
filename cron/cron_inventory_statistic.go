package cron

import (
	"start/controllers"
	"start/datasources"
)

func runReportInventoryStatisticItemJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerReportInventoryStatisticItem(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runReportInventoryStatisticItem()
}

func runReportInventoryStatisticItem() {
	db := datasources.GetDatabase()
	cStatistic := controllers.CStatisticItem{}
	cStatistic.AddItemToStatistic(db)
}
