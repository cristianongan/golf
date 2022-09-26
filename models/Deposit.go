package models

import (
	"start/constants"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Deposit struct {
	ModelId
	PartnerUid       string         `json:"partner_uid"`
	CourseUid        string         `json:"course_uid"`
	InputDate        datatypes.Date `json:"input_date"`
	CustomerUid      string         `json:"customer_uid"`
	CustomerName     string         `json:"customer_name"`
	CustomerIdentity string         `json:"customer_identity"`
	CustomerPhone    string         `json:"customer_phone"`
	CustomerType     string         `json:"customer_type"`
	PaymentType      string         `json:"payment_type"` // CC: Credit Card; CK: Chuyen khoan; TM + CK: Tien mat + chuyen khoan; TM + CC: Tien mat + credit card
	TmCurrency       string         `json:"tm_currency"`
	TmRate           float64        `json:"tm_rate"`
	TmOriginAmount   int64          `json:"tm_origin_amount"`
	TmTotalAmount    float64        `json:"tm_total_amount"`
	TcCurrency       string         `json:"tc_currency"`
	TcRate           float64        `json:"tc_rate"`
	TcOriginAmount   int64          `json:"tc_origin_amount"`
	TcTotalAmount    float64        `json:"tc_total_amount"`
	TotalAmount      float64        `json:"total_amount"`
	Note             string         `json:"note"`
}

func (item *Deposit) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *Deposit) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *Deposit) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	return db.Save(item).Error
}
