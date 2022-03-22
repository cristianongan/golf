package server

import (
	"log"
	"start/config"
	"start/datasources"
	"start/models"
)

func MigrateDb() {
	db := datasources.GetDatabase()
	if config.GetIsMigrated() {
		log.Println("migrating db")

		// ================ For Sub System ======================
		db.AutoMigrate(&models.CmsUser{})
		db.AutoMigrate(&models.CmsUserToken{})
		db.AutoMigrate(&models.Partner{})
		db.AutoMigrate(&models.Course{})
		db.AutoMigrate(&models.Todo{})
		db.AutoMigrate(&models.Buggy{})
		db.AutoMigrate(&models.BuggyDiary{})
		db.AutoMigrate(&models.MemberCard{})
		db.AutoMigrate(&models.MemberCardType{})
		db.AutoMigrate(&models.CustomerUser{})
		db.AutoMigrate(&models.Agent{})

		log.Println("migrated db")
	}
}
