package models

import (
	"start/constants"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
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

func (item *CaddieCalendar) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieCalendar) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieCalendar) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	return db.Save(item).Error
}

func (item *CaddieCalendar) Delete(db *gorm.DB) error {
	return db.Delete(item).Error
}
