package cron

import (
	"log"
	"start/datasources"
	"start/models"
	"time"

	"github.com/bsm/redislock"
)

func runResetDataMemberCardJob() {
	// Để xử lý cho chạy nhiều instance Server
	// isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerResetDataMemberCard(), 60)
	// // Ko lấy được lock, return luôn
	// if !isObtain {
	// 	return
	// }

	redisKey := datasources.GetRedisKeyLockerResetDataMemberCard()
	lock, err := datasources.GetLockerRedis().Obtain(datasources.GetCtxRedis(), redisKey, 60*time.Second, nil)
	// Ko lấy được lock, return luôn
	if err == redislock.ErrNotObtained || err != nil {
		log.Println("[CRON] runResetDataMemberCardJob Could not obtain lock", redisKey)
		return
	}

	defer lock.Release(datasources.GetCtxRedis())

	// Logic chạy cron bên dưới
	resetDataMemberCard()
}

// Reset số guest của member trong ngày
func resetDataMemberCard() {
	db := datasources.GetDatabase()
	//Lấy list member Card
	memberCardR := models.MemberCard{}
	// TODO: Udp theo lấy theo Page, sau lượng membercard lên nhiều
	err, list := memberCardR.FindAll(db)
	if err != nil {
		log.Println("resetDataMemberCard err or empty", err.Error())
		return
	}

	for _, v := range list {
		if v.TotalGuestOfDay > 0 {
			v.TotalGuestOfDay = 0
			errU := v.Update(db)
			if errU != nil {
				log.Println("resetDataMemberCard errU", errU.Error())
			}
		}
	}
}
