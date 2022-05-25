package models

import (
	"start/constants"
	"start/datasources"
	"strings"

	"github.com/pkg/errors"
)

// Quốc gia
type Nationality struct {
	Id     int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	Status string `json:"status"  gorm:"type:varchar(50)"` //ENABLE, DISABLE, TESTING
	Name   string `json:"name" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *Nationality) Create() error {
	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Nationality) Update() error {
	mydb := datasources.GetDatabase()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Nationality) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Nationality) FindAll() ([]Nationality, error) {
	db := datasources.GetDatabase().Model(Nationality{})
	list := []Nationality{}
	db.Find(&list)
	return list, db.Error
}

func (item *Nationality) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Nationality{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Nationality) FindList(page Page) ([]Nationality, int64, error) {
	db := datasources.GetDatabase().Model(Nationality{})
	list := []Nationality{}
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

func (item *Nationality) Delete() error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
