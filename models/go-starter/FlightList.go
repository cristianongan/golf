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
	FlightIndex int
}

type Map struct {
	models.ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	DateDisplay string `json:"date_display" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	CourseType  string `json:"course_type" gorm:"type:varchar(50)"`
	FlightIndex int    `json:"flight_index" gorm:"index"` // Số thức tự của flight
	Course      string `json:"course"`                    //  Sân
	Hole        int    `json:"hole"`                      // Số hố
	TimeStart   int64  `json:"time_start"`                // Thời gian bắt đầu
	TimeEnd     int64  `json:"time_end"`                  // Thời gian end
	TimeOnHole  int64  `json:"time_on_hole"`              // Thời gian end
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

	if item.FlightIndex != 0 {
		db = db.Where("flights.flight_index = ?", item.PartnerUid)
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

func (item *FlightList) FindFlightListMap(database *gorm.DB) ([]Map, error) {
	var list []Map

	// Subquery
	subQuery := database.Table("player_scores")
	if item.CourseUid != "" {
		subQuery = subQuery.Where("player_scores.course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		subQuery = subQuery.Where("player_scores.partner_uid = ?", item.PartnerUid)
	}

	subQuery.Order("player_scores.hole_index desc")

	db := database.Model(Flight{})

	db = db.Select("flights.*, tb1.course, tb1.hole, tb1.time_start, tb1.time_end")

	db = db.Joins("INNER JOIN bookings ON bookings.flight_id = flights.id")
	db = db.Joins("LEFT JOIN (?) as tb1 ON tb1.flight_id = flights.id", subQuery)

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

	db.Group("flights.id")

	db = db.Find(&list)

	return list, db.Error
}

func (item *FlightList) FindFlightMapWithIds(database *gorm.DB, ids []int64) ([]Map, error) {
	var list []Map

	// Subquery
	subQuery := database.Table("player_scores")
	if item.CourseUid != "" {
		subQuery = subQuery.Where("player_scores.course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		subQuery = subQuery.Where("player_scores.partner_uid = ?", item.PartnerUid)
	}

	subQuery.Order("player_scores.hole_index desc")

	db := database.Model(Flight{})
	db = db.Select("flights.*, tb1.course, tb1.hole, tb1.time_start, tb1.time_end")

	db = db.Joins("INNER JOIN bookings ON bookings.flight_id = flights.id")
	db = db.Joins("LEFT JOIN (?) as tb1 ON tb1.flight_id = flights.id", subQuery)

	if item.CourseUid != "" {
		db = db.Where("flights.course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("flights.partner_uid = ?", item.PartnerUid)
	}

	if len(ids) > 0 {
		db = db.Where("flights.id IN ?", ids)
	}

	db.Group("flights.id")

	db = db.Find(&list)

	return list, db.Error
}

func (item *FlightList) FindFlightWithIds(database *gorm.DB, ids []int64) ([]Flight, error) {
	var list []Flight

	db := database.Model(Flight{})
	db = db.Joins("INNER JOIN bookings ON bookings.flight_id = flights.id").Group("flights.id")

	if item.CourseUid != "" {
		db = db.Where("flights.course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("flights.partner_uid = ?", item.PartnerUid)
	}

	if len(ids) > 0 {
		db = db.Where("flights.id IN ?", ids)
	}

	db = db.Preload("Bookings", "bag_status = ?", item.BagStatus)

	db = db.Find(&list)

	return list, db.Error
}
