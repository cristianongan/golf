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
	model_payment "start/models/payment"
	model_report "start/models/report"
	model_role "start/models/role"
	model_service "start/models/service"
	model_service_restaurant_setup "start/models/service/restaurant_setup"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func MigrateDb() {
	// Db Auth
	MigrateDbAuth()

	// Db Cms
	db := datasources.GetDatabase()
	if config.GetIsMigrated() {
		log.Println("migrating db")
		MigrateDbCms(db)
		log.Println("migrated db")
	}

	if config.GetIsMigrated2() {
		db = datasources.GetDatabase().Clauses(dbresolver.Use(config.GetDbName2()))

		log.Println("migrating db2")
		MigrateDbCms(db)

		log.Println("migrated db2")
	}
}

func MigrateDbCms(db *gorm.DB) {
	// ================ For Sub System ======================
	db.AutoMigrate(&models.Buggy{})
	db.AutoMigrate(&model_gostarter.BuggyInOut{})
	db.AutoMigrate(&models.BuggyDiary{})
	db.AutoMigrate(&models.MemberCard{})
	db.AutoMigrate(&models.MemberCardType{})
	db.AutoMigrate(&models.CustomerUser{})
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
	db.AutoMigrate(&model_booking.BookingOta{})

	// ----- Payment -------
	db.AutoMigrate(&model_payment.SinglePayment{})
	db.AutoMigrate(&model_payment.SinglePaymentItem{})
	db.AutoMigrate(&model_payment.AgencyPayment{})
	db.AutoMigrate(&model_payment.AgencyPaymentItem{})

	// ---- Caddie ----
	db.AutoMigrate(&models.Caddie{})
	db.AutoMigrate(&models.CaddieNote{})
	db.AutoMigrate(&models.CaddieWorkingTime{})
	db.AutoMigrate(&models.CaddieVacationCalendar{})

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
	db.AutoMigrate(&model_gostarter.BagFlight{})
	db.AutoMigrate(&model_gostarter.CaddieBuggyInOut{})

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
	db.AutoMigrate(&models.CaddieWorkingCalendarNote{})

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

	// ------- RestuarantItem -------
	db.AutoMigrate(&models.RestaurantItem{})

	// ------- Report -------
	db.AutoMigrate(&model_report.ReportCustomerPlay{})

	// Restaurant Setup
	db.AutoMigrate(&model_service_restaurant_setup.RestaurantTableSetup{})
	db.AutoMigrate(&model_service_restaurant_setup.RestaurantTimeSetup{})

	// Maintain Schedule
	db.AutoMigrate(&models.MaintainSchedule{})

	// Notification
	db.AutoMigrate(&models.Notification{})
}

func MigrateDbAuth() {
	db := datasources.GetDatabaseAuth()
	if config.GetDbAuthIsMigrated() {
		log.Println("migrating db auth")

		db.AutoMigrate(&model_role.Role{})
		db.AutoMigrate(&model_role.Permission{})
		db.AutoMigrate(&model_role.RolePermission{})
		db.AutoMigrate(&model_role.UserRole{})
		db.AutoMigrate(&model_payment.CurrencyPaid{})
		db.AutoMigrate(&models.CmsUser{})
		db.AutoMigrate(&models.CmsUserToken{})
		db.AutoMigrate(&models.Partner{})
		db.AutoMigrate(&models.Course{})
		db.AutoMigrate(&models.CustomerType{})
	}
}
