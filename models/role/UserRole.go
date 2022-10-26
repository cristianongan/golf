package model_role

import (
	"start/models"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Role - Permission
type UserRole struct {
	Id      int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	RoleId  int64  `json:"role_id" gorm:"index"`                    // Role id
	UserUid string `json:"user_uid" gorm:"type:varchar(100);index"` // user Uid
}

// ======= CRUD ===========
func (item *UserRole) Create(db *gorm.DB) error {
	return db.Create(item).Error
}

func (item *UserRole) Update(db *gorm.DB) error {
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *UserRole) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *UserRole) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Role{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *UserRole) FindList(database *gorm.DB, page models.Page) ([]UserRole, int64, error) {
	db := database.Model(UserRole{})
	list := []UserRole{}
	total := int64(0)
	db = db.Where(item)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *UserRole) Delete(db *gorm.DB) error {
	if item.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
