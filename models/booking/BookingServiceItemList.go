package model_booking

import (
	"start/constants"

	"gorm.io/gorm"
)

type BookingServiceItemList struct {
	PartnerUid string
	CourseUid  string
	GroupCode  string
	ServiceId  string
	Name       string
	Type       string
	ItemCode   string
	FromDate   string
	ToDate     string
	BillCode   string
}

func (item *BookingServiceItemList) addFilter(db *gorm.DB) *gorm.DB {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	return db
}

func (item *BookingServiceItemList) FindAll(database *gorm.DB) (*gorm.DB, error) {
	db := database.Model(BookingServiceItem{})

	db = item.addFilter(db)

	return db, db.Error
}

func (item *BookingServiceItemList) CheckBookingCaddieExisted(database *gorm.DB) bool {
	db := database.Model(BookingServiceItem{})
	total := int64(0)

	db = item.addFilter(db)
	db = db.Where("service_type = ?", constants.CADDIE_SETTING)

	db.Count(&total)

	return total > 0
}
