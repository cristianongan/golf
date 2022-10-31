package model_payment

import (
	"gorm.io/gorm"
)

// Currency Paid
type CurrencyPaid struct {
	Id       int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	Currency string `json:"currency" gorm:"type:varchar(100);index"` // Loại tiền
	Rate     int64  `json:"rate"`                                    // Tỷ gía so với vnd đồng
}

func (item *CurrencyPaid) Create(db *gorm.DB) error {
	return db.Create(item).Error
}

func (item *CurrencyPaid) Update(mydb *gorm.DB) error {
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CurrencyPaid) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *CurrencyPaid) Count(db *gorm.DB) (int64, error) {
	db = db.Model(CurrencyPaid{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CurrencyPaid) FindAll(db *gorm.DB) ([]CurrencyPaid, error) {
	db = db.Model(CurrencyPaid{})
	list := []CurrencyPaid{}
	db.Find(&list)

	return list, db.Error
}
