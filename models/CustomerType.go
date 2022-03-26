package models

import (
	"start/datasources"

	"github.com/pkg/errors"
)

// Loại khách hàng
type CustomerType struct {
	Id       int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	Type     string `json:"type" gorm:"type:varchar(100)"`    // GUEST, MEMBER, VISITOR, FOC, AGENCY, COMPANY
	Category string `json:"category" gorm:"type:varchar(50)"` // CUSTOMER, AGENCY
}

func (item *CustomerType) Create() error {
	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CustomerType) Update() error {
	mydb := datasources.GetDatabase()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CustomerType) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CustomerType) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CustomerType{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CustomerType) FindList(page Page) ([]CustomerType, int64, error) {
	db := datasources.GetDatabase().Model(CustomerType{})
	list := []CustomerType{}
	total := int64(0)
	db = db.Where(item)
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *CustomerType) Delete() error {
	if item.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
