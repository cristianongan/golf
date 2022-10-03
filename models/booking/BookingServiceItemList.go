package model_booking

import (
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
}

func (item *BookingServiceItemList) addFilter(db *gorm.DB) *gorm.DB {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	return db
}
