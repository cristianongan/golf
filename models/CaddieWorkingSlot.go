package models

import (
	"start/constants"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CaddieWorkingSlot struct {
	ModelId
	PartnerUid string           `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string           `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	ApplyDate  string           `json:"apply_date"  gorm:"type:varchar(100);index"` // ngày áp dụng
	CaddieSlot utils.ListString `json:"caddie_slot,omitempty" gorm:"type:json"`     // Danh sách xếp nốt của caddie
}

func (item *CaddieWorkingSlot) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *CaddieWorkingSlot) IsDuplicated(db *gorm.DB) bool {
	caddieWSCheck := CaddieWorkingSlot{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		ApplyDate:  item.ApplyDate,
	}
	errFind := caddieWSCheck.FindFirst(db)
	if errFind == nil || caddieWSCheck.Id > 0 {
		return true
	}
	return false
}

func (item *CaddieWorkingSlot) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieWorkingSlot) Find(database *gorm.DB) ([]CaddieWorkingSlot, error) {
	list := []CaddieWorkingSlot{}

	db := database.Model(CaddieWorkingSlot{})

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

func (item *CaddieWorkingSlot) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	return db.Save(item).Error
}

func (item *CaddieWorkingSlot) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *CaddieWorkingSlot) DeleteBatch(db *gorm.DB) error {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ApplyDate != "" {
		db = db.Where("apply_date = ?", item.ApplyDate)
	}

	return db.Delete(item).Error
}

func (item *CaddieWorkingSlot) FindAll(database *gorm.DB) ([]CaddieWorkingSlot, error) {
	list := []CaddieWorkingSlot{}

	db := database.Model(CaddieWorkingSlot{})

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
