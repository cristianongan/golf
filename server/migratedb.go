package server

import (
	"log"
	"start/config"
	"start/datasources"
	"start/logger"
	"start/models"
	model_booking "start/models/booking"
	model_gostarter "start/models/go-starter"
	kiosk_inventory "start/models/kiosk-inventory"
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
		db.AutoMigrate(&models.Buggy{})
		db.AutoMigrate(&models.BuggyDiary{})
		db.AutoMigrate(&models.MemberCard{})
		db.AutoMigrate(&models.MemberCardType{})
		db.AutoMigrate(&models.CustomerUser{})
		db.AutoMigrate(&models.CustomerType{})
		db.AutoMigrate(&models.LockTeeTime{})
		db.AutoMigrate(&models.LockTurn{})
		db.AutoMigrate(&models.TeeTypeClose{})
		db.AutoMigrate(&model_booking.CancelBookingSetting{})

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
		db.AutoMigrate(&models.McTypeAnnualFee{})

		// ----- Booking -----
		db.AutoMigrate(&model_booking.Booking{})
		db.AutoMigrate(&model_booking.BookingWaiting{})
		db.AutoMigrate(&model_booking.BookingSetting{})
		db.AutoMigrate(&model_booking.BookingSettingGroup{})
		db.AutoMigrate(&model_booking.BookingServiceItem{})
		db.AutoMigrate(&model_booking.BookingSource{})

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

		// ------- GO --------
		db.AutoMigrate(&model_gostarter.Flight{})
		db.AutoMigrate(&model_gostarter.CaddieInOutNote{})
		db.AutoMigrate(&model_gostarter.BagFlight{})

		// ------- Service ------
		db.AutoMigrate(&model_service.Kiosk{})
		db.AutoMigrate(&model_service.Rental{})
		db.AutoMigrate(&model_service.FoodBeverage{})
		db.AutoMigrate(&model_service.FbPromotionSet{})
		db.AutoMigrate(&model_service.GroupServices{})
		db.AutoMigrate(&model_service.Proshop{})
		db.AutoMigrate(&model_service.Restaurent{})

		// ------- Deposit -------
		db.AutoMigrate(&models.Deposit{})

		// ------- Caddie Calendar -------
		db.AutoMigrate(&models.CaddieCalendar{})

		// ------- Caddie Evaluation -------
		db.AutoMigrate(&models.CaddieEvaluation{})

		// ------- Caddie Working Calendar -------
		db.AutoMigrate(&models.CaddieWorkingCalendar{})

		// ------- Caddie Group -------
		db.AutoMigrate(&models.CaddieGroup{})

		// ------- Caddie Working Schedule -------
		db.AutoMigrate(&models.CaddieWorkingSchedule{})

		// ------- System Activity Log -------
		db.AutoMigrate(&logger.ActivityLog{})

		// ------- Tranfer Card -------
		db.AutoMigrate(&models.TranferCard{})

		// ------- Caddie Fee Setting -------
		db.AutoMigrate(&models.CaddieFeeSetting{})
		db.AutoMigrate(&models.CaddieFeeSettingGroup{})
		db.AutoMigrate(&models.CaddieFee{})

		// ------- Holiday -------
		db.AutoMigrate(&models.Holiday{})

		// ------- Locker -------
		db.AutoMigrate(&models.Locker{})

		// ------- Round -------
		db.AutoMigrate(&models.Round{})

		// ------- KioskInventoryItem -------
		db.AutoMigrate(&kiosk_inventory.InventoryItem{})
		db.AutoMigrate(&kiosk_inventory.InventoryInputItem{})
		db.AutoMigrate(&kiosk_inventory.InventoryOutputItem{})
		db.AutoMigrate(&kiosk_inventory.InputInventoryBill{})
		db.AutoMigrate(&kiosk_inventory.OutputInventoryBill{})
		db.AutoMigrate(&kiosk_inventory.StatisticItem{})

		// ------- ServiceCart -------
		db.AutoMigrate(&models.ServiceCart{})

		log.Println("migrated db")
	}
}
