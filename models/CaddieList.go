package models

import (
	"start/datasources"
	"time"
)

type CaddieList struct {
	CourseUid       string
	CaddieName      string
	CaddieCode      string
	Month           string
	WorkingStatus   string
	InCurrentStatus []string
}

func (item *CaddieList) FindList(page Page) ([]Caddie, int64, error) {
	var list []Caddie
	total := int64(0)

	db := datasources.GetDatabase().Model(Caddie{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieName != "" {
		db = db.Where("name LIKE ?", "%"+item.CaddieName+"%")
	}

	if item.CaddieCode != "" {
		db = db.Where("code = ?", item.CaddieCode)
	}

	if len(item.InCurrentStatus) > 0 {
		db = db.Where("current_status IN ?", item.InCurrentStatus)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		if item.Month != "" {
			db = page.Setup(db).Preload("CaddieCalendar", "DATE_FORMAT(apply_date, '%Y-%m') = ?", item.Month).Find(&list)
		} else {
			db = page.Setup(db).Preload("CaddieCalendar", "DATE_FORMAT(apply_date, '%Y-%m') = ?", time.Now().Format("2006-01")).Find(&list)
		}
	}

	return list, total, db.Error
}

func (item CaddieList) FindFirst() (Caddie, error) {
	var result Caddie
	db := datasources.GetDatabase().Model(Caddie{})

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.CaddieName != "" {
		db = db.Where("name LIKE ?", "%"+item.CaddieName+"%")
	}

	if item.CaddieCode != "" {
		db = db.Where("code = ?", item.CaddieCode)
		item.CaddieCode = ""
	}

	if len(item.InCurrentStatus) > 0 {
		db = db.Where("current_status IN ?", item.InCurrentStatus)
	}

	err := db.First(&result).Error

	return result, err
}
