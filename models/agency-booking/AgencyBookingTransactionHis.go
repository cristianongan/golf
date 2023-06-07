package models_agency_booking

import (
	"errors"

	"gorm.io/gorm"
)

type AgencyBookingTransactionHis struct {
	Id                int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	CreatedAt         int64  `json:"created_at" gorm:"index"`
	CreatedBy         string `json:"created_by" gorm:"type:varchar(255)"`
	TransactionId     string `json:"transaction_id" gorm:"type:varchar(100);index"` // mã giao dịch ~ booking code
	TransactionStatus string `json:"transaction_status" gorm:"type:varchar(100)"`   // trạng thái đơn ~ trạng thái giao dịch
}

func (item *AgencyBookingTransactionHis) Create(db *gorm.DB) error {
	return db.Create(item).Error
}

func (item *AgencyBookingTransactionHis) FindList(db *gorm.DB) ([]AgencyBookingTransactionHis, error) {
	list := []AgencyBookingTransactionHis{}

	if item.TransactionId == "" {
		return list, errors.New("transaction id is required")
	}

	hisDb := db.Where(item)
	hisDb = db.Find(&list)

	return list, hisDb.Error
}
