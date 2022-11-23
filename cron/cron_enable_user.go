package cron

import (
	"start/constants"
	"start/datasources"
	"start/models"
)

func runEnableUserJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeySystemLogout(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
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
