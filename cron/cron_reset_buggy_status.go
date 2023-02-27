package cron

import (
	"start/constants"
	"start/datasources"
	"start/models"
)

func runResetBuggyStatusJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemResetBuggyStatus(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
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
