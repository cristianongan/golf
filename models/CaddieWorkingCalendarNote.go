package models

import (
	"start/constants"
	"start/utils"

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

func (item *CaddieWorkingCalendarNote) IsDuplicated(db *gorm.DB) bool {
	caddieNoteCheck := CaddieWorkingCalendarNote{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		ApplyDate:  item.ApplyDate,
	}
	errFind := caddieNoteCheck.FindFirst(db)
	if errFind == nil || caddieNoteCheck.Id > 0 {
		return true
	}
	return false
}

func (item *CaddieWorkingCalendarNote) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieWorkingCalendarNote) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingCalendarNote) Find(database *gorm.DB) ([]CaddieWorkingCalendarNote, error) {
	list := []CaddieWorkingCalendarNote{}

	db := database.Model(CaddieWorkingCalendarNote{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ApplyDate != "" {
		db = db.Where("apply_date = ?", item.ApplyDate)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *CaddieWorkingCalendarNote) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	return db.Save(item).Error
}

func (item *CaddieWorkingCalendarNote) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
