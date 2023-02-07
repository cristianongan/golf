package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Group Fee
type GroupFee struct {
	ModelId
	PartnerUid   string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid    string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name         string `json:"name" gorm:"type:varchar(256)"`              // Ten Group Fee
	CategoryType string `json:"category_type" gorm:"type:varchar(100)"`     // Category Type
}

func (item *GroupFee) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *GroupFee) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *GroupFee) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *GroupFee) Count(database *gorm.DB) (int64, error) {
	db := database.Model(GroupFee{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *GroupFee) FindList(database *gorm.DB, page Page) ([]GroupFee, int64, error) {
	db := database.Model(GroupFee{})
	list := []GroupFee{}
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

func (item *GroupFee) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
