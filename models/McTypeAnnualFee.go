package models

import (
	"start/constants"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Group Fee
type McTypeAnnualFee struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	McTypeId   int64  `json:"mc_type_id" gorm:"index"`                    // Member Card Type id
	Year       int    `json:"year" gorm:"index"`
	Fee        int64  `json:"fee"`
}

func (item *McTypeAnnualFee) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *McTypeAnnualFee) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *McTypeAnnualFee) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *McTypeAnnualFee) Count(database *gorm.DB) (int64, error) {
	db := database.Model(McTypeAnnualFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *McTypeAnnualFee) FindByMcTypeId(database *gorm.DB) ([]McTypeAnnualFee, error) {
	db := database.Model(McTypeAnnualFee{})
	list := []McTypeAnnualFee{}
	db = db.Where("mc_type_id = ?", item.McTypeId)

	db.Find(&list)

	return list, db.Error
}

func (item *McTypeAnnualFee) FindList(database *gorm.DB, page Page) ([]McTypeAnnualFee, int64, error) {
	db := database.Model(McTypeAnnualFee{})
	list := []McTypeAnnualFee{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *McTypeAnnualFee) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
