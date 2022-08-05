package model_booking

import (
	"fmt"
	"start/constants"
	"start/datasources"
	"start/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type BookingList struct {
	PartnerUid  string
	CourseUid   string
	BookingCode string
	BookingDate string
	CaddieUid   string
	CaddieName  string
	CaddieCode  string
	InitType    string
	AgencyId    int64
	IsAgency    string
	Status      string
	FromDate    string
	ToDate      string
	BuggyUid    string
	BuggyCode   string
	GolfBag     string
	Month       string
	IsToday     string
	BookingUid  string
	IsFlight    string
	BagStatus   string
	HaveBag     *string
	HasBuggy    string
	IsTimeOut   string
}

func addFilter(db *gorm.DB, item *BookingList) *gorm.DB {
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	if item.CaddieName != "" {
		db = db.Where("caddie_info->'$.name' LIKE ?", "%"+item.CaddieName+"%")
	}

	if item.CaddieCode != "" {
		db = db.Where("caddie_info->'$.code' = ?", item.CaddieCode)
	}

	if item.InitType != "" {
		db = db.Where("init_type = ?", item.InitType)
	}

	if item.IsAgency != "" {
		isAgency, _ := strconv.ParseInt(item.IsAgency, 10, 8)
		if isAgency == 1 {
			db = db.Where("agency_id != ?", 0)
		} else if isAgency == 0 {
			db = db.Where("agency_id = ?", 0)
		}
	}

	if item.BuggyCode != "" {
		db = db.Where("buggy_info->'$.code' = ?", item.BuggyCode)
	}

	if item.Month != "" {
		db = db.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%Y-%m') = ?", item.Month)
	}

	if item.IsToday != "" {
		isToday, _ := strconv.ParseInt(item.IsAgency, 10, 8)
		if isToday == 1 {
			currentTime := time.Now()
			db = db.Where("DATE_FORMAT(STR_TO_DATE(booking_date, '%d/%m/%Y'), '%Y-%m-%d') != ?", fmt.Sprintf("%d-%02d-%02d", currentTime.Year(), currentTime.Month(), currentTime.Day()))
		}
	}

	if item.FromDate != "" {
		db = db.Where("STR_TO_DATE(booking_date, '%d/%m/%Y') >= ?", item.FromDate)
	}

	if item.ToDate != "" {
		db = db.Where("STR_TO_DATE(booking_date, '%d/%m/%Y') <= ?", item.ToDate)
	}

	if item.GolfBag != "" {
		db = db.Where("bag = ?", item.GolfBag)
	}

	if item.BagStatus != "" {
		db = db.Where("bag_status = ?", item.BagStatus)
	}

	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}

	if item.AgencyId > 0 {
		db = db.Where("agency_id = ?", item.AgencyId)
	}

	if item.HaveBag != nil {
		if *item.HaveBag == "1" {
			db = db.Where("bag <> ?", "")
		} else {
			db = db.Where("bag = ?", "")
		}
	}
	
	if item.IsTimeOut != "" {
		isTimeOut, _ := strconv.ParseInt(item.IsTimeOut, 10, 64)
		if isTimeOut == 1 {
			db = db.Where("bag_status = ?", constants.BAG_STATUS_TIMEOUT)
		} else if isTimeOut == 0 {
			db = db.Where("bag_status <> ?", constants.BAG_STATUS_TIMEOUT).Where("bag_status <> ?", constants.BAG_STATUS_OUT)
		}
	}

	if item.IsFlight != "" {
		isFlight, _ := strconv.ParseInt(item.IsFlight, 10, 64)
		if isFlight == 1 {
			db = db.Where("flight_id <> ?", 0)
		} else if isFlight == 0 {
			db = db.Where("flight_id = ?", 0)
		}
	}

	if item.HasBuggy != "" {
		hasBuggy, _ := strconv.ParseInt(item.HasBuggy, 10, 64)
		if hasBuggy == 1 {
			db = db.Where("buggy_id <> ?", 0)
		} else if hasBuggy == 0 {
			db = db.Where("buggy_id = ?", 0)
		}
	}

	return db
}

func (item *BookingList) FindBookingList(page models.Page) ([]Booking, int64, error) {
	var list []Booking
	total := int64(0)

	db := datasources.GetDatabase().Model(Booking{})

	db = addFilter(db, item)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *BookingList) FindBookingListWithSelect(page models.Page) (*gorm.DB, int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(Booking{})

	db = addFilter(db, item)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db)
	}

	return db, total, db.Error
}

func (item *BookingList) FindFirst() (Booking, error) {
	var result Booking
	db := datasources.GetDatabase().Model(Booking{})

	if item.CaddieCode != "" {
		db = db.Where("caddie_info->'$.code' = ?", item.CaddieCode)
		item.CaddieCode = ""
	}

	if item.BookingUid != "" {
		db = db.Where("uid = ?", item.BookingUid)
		item.BookingUid = ""
	}

	if item.CaddieUid != "" {
		db = db.Where("caddie_id = ?", item.CaddieUid)
		item.CaddieUid = ""
	}

	err := db.Where(item).First(&result).Error
	return result, err
}
