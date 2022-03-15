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
		db.AutoMigrate(&models.User{})
		db.AutoMigrate(&models.Todo{})

		log.Println("migrated db")
	}
}
