package mdoel_report

import (
	"start/models"

	"gorm.io/gorm"
)

type ReportRevenueDetailList struct {
	PartnerUid string
	CourseUid  string
	FromDate   string
	ToDate     string
	GuestStyle string
}

func addFilter(db *gorm.DB, item *ReportRevenueDetailList) *gorm.DB {
	if item.PartnerUid != "" {
		db = db.Where("bookings.partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("bookings.course_uid = ?", item.CourseUid)
	}

	if item.FromDate != "" {
		db = db.Where("STR_TO_DATE(booking_date, '%d/%m/%Y') >= ?", item.FromDate)
	}

	if item.ToDate != "" {
		db = db.Where("STR_TO_DATE(booking_date, '%d/%m/%Y') <= ?", item.ToDate)
	}

	if item.GuestStyle != "" {
		db = db.Where("guest_style = ?", item.GuestStyle)
	}

	return db
}

func (item *ReportRevenueDetailList) FindBookingRevenueList(database *gorm.DB, page models.Page) (*gorm.DB, int64, error) {
	total := int64(0)

	db := database.Model(ReportRevenueDetail{})

	db = addFilter(db, item)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db)
	}

	return db, total, db.Error
}
