package model_service

import (
	"errors"
	"start/constants"
	"start/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type RestaurantTimeSetup struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	SetupType  string `json:"setup_type" gorm:"type:varchar(256)"`
	Minutes    int    `json:"minutes"`
}

func (item *RestaurantTimeSetup) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *RestaurantTimeSetup) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *RestaurantTimeSetup) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *RestaurantTimeSetup) Count(database *gorm.DB) (int64, error) {
	db := database.Model(RestaurantTimeSetup{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *RestaurantTimeSetup) FindList(database *gorm.DB) ([]RestaurantTimeSetup, int64, error) {
	db := database.Model(RestaurantTimeSetup{})
	list := []RestaurantTimeSetup{}
	total := int64(0)
	status := item.ModelId.Status

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db.Count(&total)

	db.Find(&list)

	return list, total, db.Error
}

func (item *RestaurantTimeSetup) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
