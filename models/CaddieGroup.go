package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"gorm.io/gorm"
)

type CaddieGroup struct {
	ModelId
	Name       string   `json:"name"`
	Code       string   `json:"code"`
	PartnerUid string   `json:"partner_uid"`
	CourseUid  string   `json:"course_uid"`
	Caddies    []Caddie `json:"caddies" gorm:"foreignKey:GroupId"`
}

func (item *CaddieGroup) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	// db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieGroup) ValidateCreate(db *gorm.DB) error {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.Code != "" || item.Name != "" {
		db = db.Where("code = ? OR name = ?", item.Code, item.Name)
	}

	return db.First(item).Error
}

func (item *CaddieGroup) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CaddieGroup) Delete(db *gorm.DB) error {
	return datasources.GetDatabase().Delete(item).Error
}

func (item *CaddieGroup) FindList(database *gorm.DB, page Page) ([]CaddieGroup, int64, error) {
	var list []CaddieGroup
	total := int64(0)

	db := database.Model(CaddieGroup{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Preload("Caddies", "status NOT IN (?)", constants.STATUS_DELETED).Find(&list)
	}
	return list, total, db.Error
}

func (item *CaddieGroup) FindListWithoutPage(database *gorm.DB) ([]CaddieGroup, error) {
	var list []CaddieGroup

	db := database.Model(CaddieGroup{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db = db.Preload("Caddies", "status NOT IN (?)", constants.STATUS_DELETED).Find(&list)

	return list, db.Error
}
