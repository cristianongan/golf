package cron

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"github.com/bsm/redislock"
)

func runEnableUserJob() {
	// // Để xử lý cho chạy nhiều instance Server
	// isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemLogout(), 60)
	// // Ko lấy được lock, return luôn
	// if !isObtain {
	// 	return
	// }

	redisKey := datasources.GetRedisKeySystemLogout()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 60*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[CRON] runEnableUserJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

	// Logic chạy cron bên dưới
	runEnableUser()
}

func runEnableUser() {
	user := models.CmsUser{}
	listUserLocked, _, _ := user.FindUserLocked()
	for _, v := range listUserLocked {
		redisLoginKey := datasources.GetRedisKeyUserLogin(v.UserName)
		datasources.DelCacheByKey(redisLoginKey)
		v.Status = constants.STATUS_ENABLE
		v.Update()
	}
}
