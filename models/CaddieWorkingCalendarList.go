package models

import (
	"gorm.io/gorm"
)

type CaddieWorkingCalendarList struct {
	PartnerUid string
	CourseUid  string
	ApplyDate  string
}

func (item *CaddieWorkingCalendarList) FindList(database *gorm.DB, page Page) ([]CaddieWorkingCalendar, int64, error) {
	var list []CaddieWorkingCalendar
	total := int64(0)

	db := database.Model(CaddieWorkingCalendar{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ApplyDate != "" {
		db = db.Where("DATE_FORMAT(apply_date, '%Y-%m-%d') = ?", item.ApplyDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
