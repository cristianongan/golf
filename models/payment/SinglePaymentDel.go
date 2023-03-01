package model_payment

import "gorm.io/gorm"

type SinglePaymentDel struct {
	SinglePayment
}

func (item *SinglePaymentDel) Create(db *gorm.DB) error {
	return db.Create(item).Error
}
