package models

import (
	"start/datasources"
)

type CaddieCalendarList struct {
	CourseUid  string
	CaddieUid  string
	CaddieName string
	CaddieCode string
	Month      string
	ApplyDate  string
	DayOffType string
}

func (item *CaddieCalendarList) FindList(page Page) ([]CaddieCalendar, int64, error) {
	var list []CaddieCalendar
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieCalendar{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieUid != "" {
		db = db.Where("caddie_uid = ?", item.CaddieUid)
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

	if item.ApplyDate != "" {
		db = db.Where("apply_date = ?", item.ApplyDate)
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

func (item *CaddieCalendarList) Delete() error {
	db := datasources.GetDatabase().Model(CaddieCalendar{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieUid != "" {
		db = db.Where("caddie_uid = ?", item.CaddieUid)
	}

	if item.Month != "" {
		db = db.Where("DATE_FORMAT(apply_date, '%Y-%m') = ?", item.Month)
	}

	return db.Delete(&CaddieCalendar{}).Error
}
