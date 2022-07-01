package models

import (
	"gorm.io/datatypes"
	"start/constants"
	"start/datasources"
	"time"
)

type CaddieCalendar struct {
	ModelId
	CaddieUid  string         `json:"caddie_uid" gorm:"size:256"`
	CaddieCode string         `json:"caddie_code" gorm:"size:256"`
	CaddieName string         `json:"caddie_name" gorm:"size:256"`
	PartnerUid string         `json:"partner_uid" gorm:"size:256"`
	CourseUid  string         `json:"course_uid" gorm:"size:256"`
	Title      string         `json:"title" gorm:"size:256"`
	DayOffType string         `json:"day_off_type" gorm:"size:128"`
	ApplyDate  datatypes.Date `json:"apply_date"`
	Note       string         `json:"note" gorm:"type:text"`
}

func (item *CaddieCalendar) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieCalendar) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieCalendar) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}
