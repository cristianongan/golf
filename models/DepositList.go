package models

import "start/datasources"

type DepositList struct {
	CustomerIdentity string
	CustomerPhone    string
	CustomerStyle    string
	InputDate        string
}

func (item *DepositList) FindList(page Page) ([]Deposit, int64, error) {
	var list []Deposit
	total := int64(0)

	db := datasources.GetDatabase().Model(Deposit{})

	if item.CustomerIdentity != "" {
		db = db.Where("customer_identity = ?", item.CustomerIdentity)
	}

	if item.CustomerPhone != "" {
		db = db.Where("customer_phone = ?", item.CustomerPhone)
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
