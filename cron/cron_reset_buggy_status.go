package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"github.com/bsm/redislock"
)

func runResetBuggyStatusJob() {
	// Để xử lý cho chạy nhiều instance Server
	// isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemResetBuggyStatus(), 60)
	// // Ko lấy được lock, return luôn
	// if !isObtain {
	// 	return
	// }

	redisKey := datasources.GetRedisKeySystemResetBuggyStatus()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 60*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[CRON] runResetBuggyStatusJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

	// Logic chạy cron bên dưới
	runResetBuggyStatus()
}

// Reset số guest của member trong ngày
func runResetBuggyStatus() {
	buggy := models.Buggy{}
	dbBuggy := datasources.GetDatabase()
	listBuggy, _, _ := buggy.FindListBuggyNotReady(dbBuggy)
	for _, v := range listBuggy {
		v.BuggyStatus = constants.BUGGY_CURRENT_STATUS_ACTIVE
		v.Update(dbBuggy)
	}
}
