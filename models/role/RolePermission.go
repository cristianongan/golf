package model_role

import (
	"start/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Role - Permission
type RolePermission struct {
	Id            int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	RoleId        int64  `json:"role_id" gorm:"index"`                          // Role id
	PermissionUid string `json:"permission_uid" gorm:"type:varchar(256);index"` // Permission Uid
}

// ======= CRUD ===========
func (item *RolePermission) Create(db *gorm.DB) error {
	return db.Create(item).Error
}

func (item *RolePermission) Update(db *gorm.DB) error {
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *RolePermission) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *RolePermission) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Role{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *RolePermission) FindList(database *gorm.DB, page models.Page) ([]RolePermission, int64, error) {
	db := database.Model(RolePermission{})
	list := []RolePermission{}
	total := int64(0)
	db = db.Where(item)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *RolePermission) Delete(db *gorm.DB) error {
	if item.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
