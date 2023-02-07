package models

import (
	"errors"
	"start/constants"
	"start/utils"
	"strings"

	"gorm.io/gorm"
)

// System config job
type SystemConfigJob struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *SystemConfigJob) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *SystemConfigJob) Update(db *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *SystemConfigJob) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *SystemConfigJob) Count(database *gorm.DB) (int64, error) {
	db := database.Model(SystemConfigJob{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *SystemConfigJob) FindList(database *gorm.DB, page Page) ([]SystemConfigJob, int64, error) {
	db := database.Model(SystemConfigJob{})
	list := []SystemConfigJob{}
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

func (item *SystemConfigJob) Delete(db *gorm.DB) error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
