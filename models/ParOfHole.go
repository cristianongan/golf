package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Bảng phí
type ParOfHole struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	CourseType string `json:"course_tye" gorm:"type:varchar(50);index"`   // Loại sân
	Course     string `json:"course" gorm:"type:varchar(50);index"`       //  Sân
	Hole       int    `json:"hole" gorm:"index"`                          // Số hố
	Par        int    `json:"par"`                                        // Số lần chạm gậy
	Minute     int    `json:"minute"`                                     // Số phút
}

func (item *ParOfHole) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *ParOfHole) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *ParOfHole) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *ParOfHole) FindList(database *gorm.DB, page Page) ([]ParOfHole, int64, error) {
	db := database.Model(ParOfHole{})
	list := []ParOfHole{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	db.Order("course, hole")

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *ParOfHole) FindListAll(database *gorm.DB) ([]ParOfHole, error) {
	db := database.Model(ParOfHole{})
	list := []ParOfHole{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	db.Order("course, hole")

	db = db.Find(&list)

	return list, db.Error
}

func (item *ParOfHole) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *ParOfHole) DeleteBatch(db *gorm.DB) error {
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.Course != "" {
		db = db.Where("course = ?", item.Course)
	}

	return db.Delete(item).Error
}
