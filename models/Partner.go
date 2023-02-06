package models

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"
)

// Hãng Golf
type Partner struct {
	Model
	Name string `json:"name" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *Partner) Create() error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabaseAuth()
	return db.Create(item).Error
}

func (item *Partner) Update() error {
	mydb := datasources.GetDatabaseAuth()
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Partner) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *Partner) Count() (int64, error) {
	db := datasources.GetDatabaseAuth().Model(Partner{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Partner) FindList(page Page) ([]Partner, int64, error) {
	db := datasources.GetDatabaseAuth().Model(Partner{})
	list := []Partner{}
	total := int64(0)
	status := item.Status
	item.Status = ""
	db = db.Where(item)

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}

	if item.Name != "" {
		db = db.Where("name like ?", "%"+item.Name+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Partner) Delete() error {
	if item.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabaseAuth().Delete(item).Error
}
