package models

import (
	"gorm.io/datatypes"
	"start/constants"
	"start/datasources"
	"time"
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

func (item *CaddieWorkingCalendar) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieWorkingCalendar) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingCalendar) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}
