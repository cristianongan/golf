package model_role

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"

	"github.com/pkg/errors"
)

// Permission
type Permission struct {
	Uid         string                 `gorm:"primary_key" sql:"not null;" json:"uid"`
	Status      string                 `json:"status" gorm:"index;type:varchar(50)"`    //ENABLE, DISABLE, TESTING
	Name        string                 `json:"name" gorm:"type:varchar(200)"`           // Name Permission
	Category    string                 `json:"category" gorm:"type:varchar(100);index"` // Category
	Description string                 `json:"description" gorm:"type:varchar(200)"`    // description
	Resources   ListPermissionResource `json:"resources,omitempty" gorm:"type:json"`    // Permission Resources
}

type PermissionResource struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type ListPermissionResource []PermissionResource

func (item *ListPermissionResource) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListPermissionResource) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ======= CRUD ===========
func (item *Permission) Create() error {
	db := datasources.GetDatabaseAuth()
	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *Permission) Update() error {
	db := datasources.GetDatabaseAuth()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Permission) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *Permission) Count() (int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(Permission{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Permission) FindList(page models.Page) ([]Permission, int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(Permission{})
	list := []Permission{}
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

func (item *Permission) Delete() error {
	db := datasources.GetDatabaseAuth()
	if item.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
