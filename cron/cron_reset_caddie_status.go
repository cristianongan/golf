package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"github.com/bsm/redislock"
)

func runResetCaddieStatusJob() {
	// Để xử lý cho chạy nhiều instance Server
	// isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemResetCaddieStatus(), 60)
	// // Ko lấy được lock, return luôn
	// if !isObtain {
	// 	return
	// }
	redisKey := datasources.GetRedisKeySystemResetCaddieStatus()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 60*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[CRON] runResetCaddieStatusJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

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
	log.Println("[CRON] runResetCaddieStatus len ", len(listCaddie))
	for _, v := range listCaddie {
		log.Println("[CRON] runResetCaddieStatus code ", v.Code)
		if v.CurrentStatus != constants.CADDIE_CURRENT_STATUS_READY {
			log.Println("[CRON] runResetCaddieStatus CurrentStatus ", v.CurrentStatus)
		}
		v.CurrentStatus = constants.CADDIE_CURRENT_STATUS_READY
		v.CurrentRound = 0
		errUdp := v.Update(dbCaddie)
		if errUdp != nil {
			log.Println("[CRON] runResetCaddieStatus errUdp ", errUdp.Error())
		}
	}
}
