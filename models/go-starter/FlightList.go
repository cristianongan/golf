package model_gostarter

import (
	"start/datasources"
	"start/models"
	// "gorm.io/gorm"
)

type FlightList struct {
	PartnerUid           string
	CourseUid            string
	BookingDate          string
	GolfBag              string
	CaddieName           string
	PlayerName           string
	CaddieCode           string
	CustomerName         string
	PeopleNumberInFlight *int
}

func (item *FlightList) FindFlightList(page models.Page) ([]Flight, int64, error) {
	var list []Flight
	total := int64(0)

	db := datasources.GetDatabase().Model(Flight{})
	db = db.Joins("INNER JOIN bookings ON bookings.flight_id = flights.id").Group("flights.id")

	if item.GolfBag != "" {
		db = db.Where("bookings.bag = ?", item.GolfBag)
	}

	if item.CustomerName != "" {
		db = db.Where("bookings.customer_name LIKE ?", "%"+item.CustomerName+"%")
	}

	if item.CaddieName != "" {
		db = db.Where("bookings.caddie_info->'$.name' LIKE ?", "%"+item.CaddieName+"%")
	}

	if item.CaddieCode != "" {
		db = db.Where("bookings.caddie_info->'$.code' = ?", item.CaddieCode)
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

	db.Count(&total)
	db = db.Preload("Bookings").Preload("Bookings.CaddieInOut")

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	if item.PeopleNumberInFlight != nil {
		listResponse := []Flight{}
		for _, data := range list {
			if len(data.Bookings) == *item.PeopleNumberInFlight {
				listResponse = append(listResponse, data)
			}
		}
		return listResponse, total, db.Error
	}

	return list, total, db.Error
}
