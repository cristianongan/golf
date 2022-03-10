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
		log.Println("migrate db")

		// ================ For Sub System ======================
		db.AutoMigrate(&models.User{})

		log.Println("Migrate db")
	}
}
