package models

import (
	"gorm.io/datatypes"
	"start/datasources"
)

type CaddieCalendarList struct {
	CourseUid  string
	CaddieName string
	CaddieCode string
	Month      string
	ApplyDate  datatypes.Date
}

func (item *CaddieCalendarList) FindList(page Page) ([]CaddieCalendar, int64, error) {
	var list []CaddieCalendar
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieCalendar{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieName != "" {
		db = db.Where("caddie_name = ?", item.CaddieName)
	}

	if item.CaddieCode != "" {
		db = db.Where("caddie_code = ?", item.CaddieCode)
	}

	if item.Month != "" {
		db = db.Where("DATE_FORMAT(apply_date, '%Y-%m') = ?", item.Month)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CaddieCalendarList) FindFirst() (CaddieCalendar, error) {
	var result CaddieCalendar
	db := datasources.GetDatabase().Model(CaddieCalendar{})
	err := db.Where(item).First(&result).Error
	return result, err
}
