package main

import (
	"log"
	"start/config"
	"start/datasources"
)

func main() {
	config.ReadConfigFile("staging")
	datasources.MySqlConnect()
	//db := datasources.GetDatabase()
	log.Println("RUN MIGRATE DB LOCAL")

	// Caddie
	//db.AutoMigrate(&models.CaddieEvaluation{})
	//log.Println("CaddieEvaluation")
	//db.AutoMigrate(&models.CaddieCalendar{})
	//log.Println("CaddieCalendar")
	//db.AutoMigrate(&models.CaddieWorkingCalendar{})
	//log.Println("CaddieWorkingCalendar")
}
