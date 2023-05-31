package models

import (
	"errors"
	"start/constants"
	"start/utils"
	"strings"

	"gorm.io/gorm"
)

type ConfigTimeNoti struct {
	ModelId
	CourseUid        string `json:"course_uid" gorm:"size:256" binding:"required"`
	PartnerUid       string `json:"partner_uid" gorm:"size:256" binding:"required"`
	TimeIntervalName string `json:"time_interval_name" binding:"required" grom:"type:varchar(100)"`
	TimeIntervalType string `json:"time_interval_type" binding:"required"`
	ColorCode        string `json:"color_code" grom:"type:varchar(50)"`
	Description      string `json:"description" grom:"varchar(255)"`
	FirstMilestone   int    `json:"first_milestone"`
	SecondMilestone  int    `json:"second_milestone"`
}

func (item *ConfigTimeNoti) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	return db.Create(item).Error
}

func (item *ConfigTimeNoti) Update(db *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *ConfigTimeNoti) FindList(page Page, db *gorm.DB) ([]ConfigTimeNoti, int64, error) {
	db = db.Model(&ConfigTimeNoti{})

	list := []ConfigTimeNoti{}
	total := int64(0)
	status := item.Status

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}

	if item.PartnerUid != "" && item.PartnerUid != constants.ROOT_PARTNER_UID {
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

func (item *ConfigTimeNoti) FindAll(db *gorm.DB) ([]ConfigTimeNoti, int64, error) {
	db = db.Model(&ConfigTimeNoti{})

	list := []ConfigTimeNoti{}
	total := int64(0)
	status := item.Status

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}

	if item.PartnerUid != "" && item.PartnerUid != constants.ROOT_PARTNER_UID {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db.Count(&total)

	db = db.Find(&list)

	return list, total, db.Error
}

func (item *ConfigTimeNoti) Delete(db *gorm.DB) error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *ConfigTimeNoti) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(&item).Error
}

func (item *ConfigTimeNoti) FindFirstExclude(idExcludes []int64, db *gorm.DB) error {
	return db.Where(item).Where(" id NOT IN (?) ", idExcludes).First(&item).Error
}
