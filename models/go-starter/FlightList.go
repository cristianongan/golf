package model_gostarter

import (
	"start/datasources"
	"start/models"
)

type FlightList struct {
	BookingDate string
}

func (item *FlightList) FindFlightList(page models.Page) ([]Flight, int64, error) {
	var list []Flight
	total := int64(0)

	db := datasources.GetDatabase().Model(Flight{})

	if item.BookingDate != "" {
		db = db.Where("date_display = ?", item.BookingDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Preload("Bookings").Find(&list)
	}

	return list, total, db.Error
}
