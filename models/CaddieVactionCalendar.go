package models

import (
	"start/constants"
	"start/utils"
	"time"

	"gorm.io/gorm"
)

type CaddieVacationCalendar struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	CaddieId      int64  `json:"caddie_id" gorm:"index"`
	CaddieCode    string `json:"caddie_code" gorm:"type:varchar(100);index"`
	CaddieName    string `json:"caddie_name" gorm:"type:varchar(256)"`
	Title         string `json:"title" gorm:"type:varchar(100)"`
	Color         string `json:"color" gorm:"type:varchar(100)"`
	DateFrom      int64  `json:"date_from"`
	DateTo        int64  `json:"date_to"`
	MonthFrom     int    `json:"month_from"`
	MonthTo       int    `json:"month_to"`
	NumberDayOff  int    `json:"number_day_off"`
	Note          string `json:"note" gorm:"type:varchar(256)"`
	ApproveStatus string `json:"approve_status"`
	ApproveTime   int64  `json:"approve_time"`
	UserApprove   string `json:"user_approve"`
}

func (item *CaddieVacationCalendar) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieVacationCalendar) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieVacationCalendar) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	return db.Save(item).Error
}

func (item *CaddieVacationCalendar) Delete(db *gorm.DB) error {
	return db.Delete(item).Error
}

func (item *CaddieVacationCalendar) FindAll(database *gorm.DB) ([]CaddieVacationCalendar, error) {
	var list []CaddieVacationCalendar

	db := database.Model(CaddieVacationCalendar{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CaddieId != 0 {
		db = db.Where("caddie_id = ?", item.CaddieId)
	}

	if item.MonthFrom != 0 {
		db = db.Where("month_from <= ?", item.MonthFrom)
		db = db.Where("month_to >= ?", item.MonthFrom)
	}

	if item.DateFrom != 0 {
		db = db.Where("date_from <= ?", item.DateFrom)
		db = db.Where("date_to >= ?", item.DateFrom)
	}

	db = db.Find(&list)

	return list, db.Error
}

func (item *CaddieVacationCalendar) FindAllWithDate(database *gorm.DB, typeWork string, date time.Time) ([]CaddieVacationCalendar, error) {
	var list []CaddieVacationCalendar

	db := database.Model(CaddieVacationCalendar{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.ApproveStatus != "" {
		db = db.Where("approve_status = ?", item.ApproveStatus)
	}

	if typeWork == "LEAVE" {
		db = db.Where("date_from <= ?", date.Unix())
		db = db.Where("date_to >= ?", date.Unix())
	}

	if typeWork == "WORK" {
		db = db.Where("date_to <= ?", date.Unix())
		db = db.Where("date_to >= ?", date.AddDate(0, 0, -1).Unix())
	}

	db.Order("approve_time asc")

	db.Group("caddie_code")

	db = db.Find(&list)

	return list, db.Error
}
