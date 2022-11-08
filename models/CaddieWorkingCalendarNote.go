package models

import (
	"start/constants"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CaddieWorkingCalendarNote struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	ApplyDate  string `json:"apply_date"  gorm:"type:varchar(100)"`       // ngày áp dụng
	Note       string `json:"note" gorm:"type:varchar(256)"`              // Note booking caddie
}

func (item *CaddieWorkingCalendarNote) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieWorkingCalendarNote) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingCalendarNote) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	return db.Save(item).Error
}

func (item *CaddieWorkingCalendarNote) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
