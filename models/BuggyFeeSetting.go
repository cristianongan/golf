package models

import (
	"start/constants"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BuggyFeeSetting struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	TypeName   string `json:"type_name" gorm:"type:varchar(100)"`
}

// ======= CRUD ===========
func (item *BuggyFeeSetting) Create(db *gorm.DB) error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BuggyFeeSetting) Update(db *gorm.DB) error {
	item.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BuggyFeeSetting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BuggyFeeSetting) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BuggyFeeSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BuggyFeeSetting) FindList(database *gorm.DB, page Page) ([]BuggyFeeSetting, int64, error) {
	db := database.Model(BuggyFeeSetting{})
	list := []BuggyFeeSetting{}
	total := int64(0)
	if item.Status != "" {
		db = db.Where("status IN (?)", strings.Split(item.Status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	db.Count(&total)
	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *BuggyFeeSetting) FindAll(database *gorm.DB) ([]BuggyFeeSetting, int64, error) {
	db := database.Model(BuggyFeeSetting{})
	list := []BuggyFeeSetting{}
	total := int64(0)
	if item.Status != "" {
		db = db.Where("status IN (?)", strings.Split(item.Status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	db.Count(&total)

	db = db.Find(&list)
	return list, total, db.Error
}

func (item *BuggyFeeSetting) Delete(database *gorm.DB) error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return database.Delete(item).Error
}
