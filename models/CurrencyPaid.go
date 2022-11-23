package models

import (
	"start/datasources"
)

// Currency Paid
type CurrencyPaid struct {
	Id       int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	Currency string `json:"currency" gorm:"type:varchar(100);index"` // Loại tiền
	Rate     int64  `json:"rate"`                                    // Tỷ gía so với vnd đồng
}

func (item *CurrencyPaid) Create() error {
	db := datasources.GetDatabaseAuth()
	return db.Create(item).Error
}

func (item *CurrencyPaid) Update() error {
	db := datasources.GetDatabaseAuth()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CurrencyPaid) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *CurrencyPaid) Count() (int64, error) {
	db := datasources.GetDatabaseAuth()
	db = db.Model(CurrencyPaid{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CurrencyPaid) FindAll() ([]CurrencyPaid, error) {
	db := datasources.GetDatabaseAuth()
	db = db.Model(CurrencyPaid{})
	list := []CurrencyPaid{}
	db.Find(&list)

	return list, db.Error
}
