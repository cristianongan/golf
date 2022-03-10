package models

import (
	"errors"
	"start/datasources"
)

type UserTest struct {
	Phone     string `gorm:"primary_key" sql:"not null;" json:"phone"`
	Name      string `json:"name" sql:"size:100"`
	Signature string `json:"signature"`
}

func (item *UserTest) Create() error {
	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *UserTest) Update() error {
	mydb := datasources.GetDatabase()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *UserTest) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *UserTest) Count() (int64, error) {
	db := datasources.GetDatabase().Model(UserTest{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *UserTest) Delete() error {
	if item.Phone == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
