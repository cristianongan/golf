package server

import (
	"log"
	"start/config"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	model_service "start/models/service"
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
		db.AutoMigrate(&models.CustomerType{})

		// ----- Agency ------
		db.AutoMigrate(&models.Agency{})
		db.AutoMigrate(&models.AgencySpecialPrice{})

		// ----- Fee -------
		db.AutoMigrate(&models.TablePrice{})
		db.AutoMigrate(&models.GolfFee{})
		db.AutoMigrate(&models.GroupFee{})
		db.AutoMigrate(&models.HolePriceFormula{})
		db.AutoMigrate(&models.AnnualFee{})
		db.AutoMigrate(&models.AnnualFeePay{})

		// ----- Booking -----
		db.AutoMigrate(&model_booking.Booking{})
		db.AutoMigrate(&model_booking.BookingSetting{})
		db.AutoMigrate(&model_booking.BookingSettingGroup{})

		// ---- Caddie ----
		db.AutoMigrate(&models.Caddie{})
		db.AutoMigrate(&models.CaddieNote{})
		db.AutoMigrate(&models.CaddieWorkingTime{})

		// ---- Bag Note ----
		db.AutoMigrate(&models.BagsNote{})

		// ------ System ------
		db.AutoMigrate(&models.SystemConfigJob{})
		db.AutoMigrate(&models.SystemConfigPosition{})
		db.AutoMigrate(&models.Nationality{})
		db.AutoMigrate(&models.CompanyType{})
		db.AutoMigrate(&models.Company{})

		// ------- Service ------
		db.AutoMigrate(&model_service.Kiosk{})
		db.AutoMigrate(&model_service.Rental{})
		db.AutoMigrate(&model_service.Proshop{})
		db.AutoMigrate(&model_service.Restaurent{})

		log.Println("migrated db")
	}
}
