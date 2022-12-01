package model_role

import (
	"log"
	"start/datasources"
	"start/models"

	"github.com/pkg/errors"
)

// Role - Permission
type RolePermission struct {
	Id            int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	RoleId        int64  `json:"role_id" gorm:"index"`                          // Role id
	PermissionUid string `json:"permission_uid" gorm:"type:varchar(256);index"` // Permission Uid
}

// ======= CRUD ===========
func (item *RolePermission) Create() error {
	db := datasources.GetDatabaseAuth()
	return db.Create(item).Error
}

func (item *RolePermission) BatchInsert(list []RolePermission) error {
	database := datasources.GetDatabaseAuth()
	db := database.Model(RolePermission{})
	var err error
	err = db.Create(&list).Error

	if err != nil {
		log.Println("RolePermission batch insert err: ", err.Error())
	}
	return err
}

func (item *RolePermission) Update() error {
	db := datasources.GetDatabaseAuth()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *RolePermission) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *RolePermission) Count() (int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(RolePermission{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *RolePermission) FindAll() ([]RolePermission, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(RolePermission{})
	list := []RolePermission{}

	db = db.Where(item)

	db.Find(&list)
	return list, db.Error
}

func (item *RolePermission) FindAllPermission() ([]Permission, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(RolePermission{})
	list := []Permission{}
	db = db.Where(item)
	db = db.Joins("JOIN permissions ON role_permissions.permission_uid = permissions.uid")
	db = db.Select("permissions.uid, permissions.status,permissions.category,permissions.description,permissions.resources")

	db.Debug().Find(&list)
	return list, db.Error
}

func (item *RolePermission) FindList(page models.Page) ([]RolePermission, int64, error) {
	database := datasources.GetDatabaseAuth()
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

func (item *RolePermission) Delete() error {
	db := datasources.GetDatabaseAuth()
	if item.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *RolePermission) DeleteList(list []RolePermission) error {
	db := datasources.GetDatabaseAuth()
	return db.Delete(list).Error
}
