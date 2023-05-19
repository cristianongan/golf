package model_gostarter

import (
	"start/models"

	"gorm.io/gorm"
	// "gorm.io/gorm"
)

type FlightList struct {
	PartnerUid  string
	CourseUid   string
	BookingDate string
	GolfBag     string
	CaddieName  string
	PlayerName  string
	CaddieCode  string
	BagStatus   string
}

func (item *FlightList) FindFlightList(database *gorm.DB, page models.Page) ([]Flight, int64, error) {
	var list []Flight
	total := int64(0)

	db := database.Model(Flight{})
	db = db.Joins("INNER JOIN bookings ON bookings.flight_id = flights.id").Group("flights.id")

	if item.GolfBag != "" {
		db = db.Where("bookings.bag COLLATE utf8mb4_general_ci LIKE ?", "%"+item.GolfBag+"%")
	}

	if item.PlayerName != "" {
		db = db.Where("bookings.customer_name COLLATE utf8mb4_general_ci LIKE ?", "%"+item.PlayerName+"%")
	}

	if item.CaddieName != "" {
		db = db.Where("bookings.caddie_info->'$.name' COLLATE utf8mb4_general_ci LIKE ?", "%"+item.CaddieName+"%")
	}

	if item.CaddieCode != "" {
		db = db.Where("bookings.caddie_info->'$.code' = ?", item.CaddieCode)
	}

	if item.BagStatus != "" {
		db = db.Where("bookings.bag_status = ?", item.BagStatus)
	}

	if item.BookingDate != "" {
		db = db.Where("flights.date_display = ?", item.BookingDate)
	}

	if item.CourseUid != "" {
		db = db.Where("flights.course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("flights.partner_uid = ?", item.PartnerUid)
	}

	if page.SortDir != "" {
		if page.SortDir == "asc" {
			db = db.Order("flights.created_at asc")
		}
		if page.SortDir == "desc" {
			db = db.Order("flights.created_at desc")
		}
	}

	db.Count(&total)
	if item.BagStatus != "" {
		db = db.Preload("Bookings", "bag_status = ?", item.BagStatus).Preload("Bookings.CaddieBuggyInOut")
	} else {
		db = db.Preload("Bookings").Preload("Bookings.CaddieBuggyInOut")
	}

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
