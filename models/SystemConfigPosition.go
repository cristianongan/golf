package models

import (
	"errors"
	"start/constants"
	"strings"
	"time"

	"gorm.io/gorm"
)

// System config Position
type SystemConfigPosition struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`
}

// ========== CRUD ===========
func (item *SystemConfigPosition) Create(db *gorm.DB) error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}
	
	return db.Create(item).Error
}

func (item *SystemConfigPosition) Update(db *gorm.DB) error {
	item.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *SystemConfigPosition) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *SystemConfigPosition) Count(database *gorm.DB) (int64, error) {
	db := database.Model(SystemConfigPosition{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *SystemConfigPosition) FindList(database *gorm.DB,page Page) ([]SystemConfigPosition, int64, error) {
	db := database.Model(SystemConfigPosition{})
	list := []SystemConfigPosition{}
	total := int64(0)
	status := item.Status
	item.Status = ""
	db = db.Where(item)

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *SystemConfigPosition) Delete(db *gorm.DB) error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
