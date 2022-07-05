package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

type GroupServices struct {
	models.ModelId
	GroupCode   string `json:"group_code" gorm:"type:varchar(100)"`   // Mã Group
	GroupName   string `json:"group_name" gorm:"type:varchar(256)"`   // Tên Group
	DetailGroup string `json:"detail_group" gorm:"type:varchar(256)"` // Tên Group
	Type        string `json:"type" gorm:"type:varchar(100)"`         // Loại service, kiosk, proshop.
}

type GroupServicesResponse struct {
	GroupCode   string `json:"group_code"`
	GroupName   string `json:"group_name"`
	Type        string `json:"type"`
	DetailGroup string `json:"detail_group"`
}

func (item *GroupServices) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *GroupServices) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *GroupServices) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *GroupServices) Count() (int64, error) {
	db := datasources.GetDatabase().Model(GroupServices{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *GroupServices) FindList(page models.Page) ([]GroupServicesResponse, int64, error) {
	db := datasources.GetDatabase().Model(GroupServices{})
	list := []GroupServicesResponse{}
	total := int64(0)
	db = db.Where(item)
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *GroupServices) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
