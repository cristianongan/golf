package cron

import (
	"log"
	"start/datasources"
	"start/models"
)

func runResetDataMemberCardJob() {
	// Để xử lý cho chạy nhiều instance Server
	isObtain := datasources.GetLockerRedisObtainWith(datasources.GetRedisKeyLockerResetDataMemberCard(), 60)
	// Ko lấy được lock, return luôn
	if !isObtain {
		return
	}
	// Logic chạy cron bên dưới
	resetDataMemberCard()
}

// Reset số guest của member trong ngày
func resetDataMemberCard() {
	//Lấy list member Card
	memberCardR := models.MemberCard{}
	// TODO: Udp theo lấy theo Page, sau lượng membercard lên nhiều
	err, list := memberCardR.FindAll()
	if err != nil {
		log.Println("resetDataMemberCard err or empty", err.Error())
		return
	}

	for _, v := range list {
		if v.TotalGuestOfDay > 0 {
			v.TotalGuestOfDay = 0
			errU := v.Update()
			if errU != nil {
				log.Println("resetDataMemberCard errU", errU.Error())
			}
		}
	}
}
