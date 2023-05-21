package cron

import (
	"log"
	"start/controllers"
	"start/datasources"
	"time"

	"github.com/bsm/redislock"
)

func runReportInventoryStatisticItemJob() {
	// Để xử lý cho chạy nhiều instance Server
	// isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerReportInventoryStatisticItem(), 60)
	// // Ko lấy được lock, return luôn
	// if !isObtain {
	// 	return
	// }

	redisKey := datasources.GetRedisKeyLockerReportInventoryStatisticItem()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 60*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[CRON] runReportInventoryStatisticItemJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

	// Logic chạy cron bên dưới
	runReportInventoryStatisticItem()
}

func runReportInventoryStatisticItem() {
	db := datasources.GetDatabase()
	cStatistic := controllers.CStatisticItem{}
	cStatistic.AddItemToStatistic(db)
}
