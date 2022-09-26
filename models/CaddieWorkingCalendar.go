package models

import (
	"start/constants"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CaddieWorkingCalendar struct {
	ModelId
	CaddieUid    string         `json:"caddie_uid" gorm:"size:256"`
	CaddieCode   string         `json:"caddie_code" gorm:"size:256"`
	PartnerUid   string         `json:"partner_uid" gorm:"size:256"`
	CourseUid    string         `json:"course_uid" gorm:"size:256"`
	CaddieLabel  string         `json:"caddie_label" gorm:"size:128"`
	CaddieColumn int            `json:"caddie_column" gorm:"size:2"`
	CaddieRow    string         `json:"caddie_row" gorm:"size:128"`
	RowTime      datatypes.Time `json:"row_time"`
	ApplyDate    datatypes.Date `json:"apply_date"`
}

func (item *CaddieWorkingCalendar) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieWorkingCalendar) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingCalendar) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	return db.Save(item).Error
}
