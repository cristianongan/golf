package model_service

import (
	"errors"
	"start/constants"
	"start/models"
	"start/utils"
	"strings"

	"gorm.io/gorm"
)

type RestaurantTableSetup struct {
	models.ModelId
	PartnerUid       string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid        string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	NumberOfFloor    int    `json:"number_of_floor"`
	NameFloor        string `json:"name_floor" gorm:"type:varchar(256)"`
	NumberOfTables   int    `json:"number_of_tables"`
	MaxPersonInTable int    `json:"max_person_in_table"`
}

func (item *RestaurantTableSetup) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *RestaurantTableSetup) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *RestaurantTableSetup) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *RestaurantTableSetup) Count(database *gorm.DB) (int64, error) {
	db := database.Model(RestaurantTableSetup{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *RestaurantTableSetup) FindList(database *gorm.DB) ([]RestaurantTableSetup, int64, error) {
	db := database.Model(RestaurantTableSetup{})
	list := []RestaurantTableSetup{}
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

func (item *RestaurantTableSetup) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
