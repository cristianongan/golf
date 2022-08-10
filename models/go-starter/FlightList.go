package model_gostarter

import (
	"start/datasources"
	"start/models"
)

type FlightList struct {
	PartnerUid           string
	CourseUid            string
	BookingDate          string
	PeopleNumberInFlight *int
}

func (item *FlightList) FindFlightList(page models.Page) ([]Flight, error) {
	var list []Flight
	total := int64(0)

	db := datasources.GetDatabase().Model(Flight{})

	if item.BookingDate != "" {
		db = db.Where("date_display = ?", item.BookingDate)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Preload("Bookings").Find(&list)
	}

	if item.PeopleNumberInFlight != nil {
		listResponse := []Flight{}
		for _, data := range list {
			if len(data.Bookings) == *item.PeopleNumberInFlight {
				listResponse = append(listResponse, data)
			}
		}
		return listResponse, db.Error
	}

	return list, db.Error
}
