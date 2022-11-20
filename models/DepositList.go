package models

import (
	"gorm.io/gorm"
)

type DepositList struct {
	CustomerIdentity string
	CustomerPhone    string
	CustomerStyle    string
	InputDate        string
}

func (item *DepositList) FindList(database *gorm.DB, page Page) ([]Deposit, int64, error) {
	var list []Deposit
	total := int64(0)

	db := database.Model(Deposit{})

	if item.CustomerIdentity != "" {
		db = db.Where("customer_identity LIKE  ?", "%"+item.CustomerIdentity+"%")
	}

	if item.CustomerPhone != "" {
		db = db.Where("customer_phone LIKE ?", "%"+item.CustomerPhone+"%")
	}

	if item.CustomerStyle != "" {
		db = db.Where("customer_style = ?", item.CustomerStyle)
	}

	if item.InputDate != "" {
		db = db.Where("input_date = ?", item.InputDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
