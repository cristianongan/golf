package cron

import (
	"start/constants"
	"start/datasources"
	"start/models"
)

func runResetCaddieStatusJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemResetCaddieStatus(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	runResetCaddieStatus()
}

// Reset số guest của member trong ngày
func runResetCaddieStatus() {
	caddie := models.CaddieList{}
	dbCaddie := datasources.GetDatabase()
	listCaddie, _, _ := caddie.FindAllCaddieList(dbCaddie) // Lấy ra caddie trong ngày làm việc
	/*
		Reset het trang thai cua nhung thang do
	*/
	for _, v := range listCaddie {
		v.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
		v.CurrentRound = 0
		v.Update(dbCaddie)
	}
}
