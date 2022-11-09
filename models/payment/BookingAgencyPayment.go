package model_payment

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Booking Agency Payment
type BookingAgencyPayment struct {
	models.ModelId
	PartnerUid  string                          `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid   string                          `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingCode string                          `json:"booking_code" gorm:"type:varchar(100);index"` // Booking code
	AgencyId    int64                           `json:"agency_id" gorm:"index"`                      // agency id
	TotalPaid   int64                           `json:"total_paid"`                                  // Số tiền thanh toán
	Players     ListBookingAgencyPaymentPlayer  `json:"players" gorm:"type:json"`                    // Booking uids dc agency thanh toán, có options caddie id
	Fees        ListBookingAgencyPaymentFeeData `json:"fees" gorm:"type:json"`                       // chi tiết Fee
}

// Player
type BookingAgencyPaymentPlayer struct {
	BookingUid string `json:"booking_uid"`
	CaddieId   string `json:"caddie_id"` // id caddie, nếu player có chọn caddie thì có
}

type ListBookingAgencyPaymentPlayer []BookingAgencyPaymentPlayer

func (item *ListBookingAgencyPaymentPlayer) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingAgencyPaymentPlayer) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Fees Data

type ListBookingAgencyPaymentFeeData []BookingAgencyPaymentFeeData

func (item *ListBookingAgencyPaymentFeeData) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingAgencyPaymentFeeData) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type BookingAgencyPaymentFeeData struct {
	Fee          int64  `json:"fee"`
	TotalFee     int64  `json:"total_fee"`
	NumberPeople int    `json:"number_people"`
	Name         string `json:"name"`
}

func (item *BookingAgencyPayment) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingAgencyPayment) Update(mydb *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingAgencyPayment) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingAgencyPayment) Count(db *gorm.DB) (int64, error) {
	db = db.Model(BookingAgencyPayment{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingAgencyPayment) FindAll(db *gorm.DB) ([]BookingAgencyPayment, error) {
	db = db.Model(BookingAgencyPayment{})
	list := []BookingAgencyPayment{}
	status := constants.STATUS_ENABLE
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *BookingAgencyPayment) FindList(db *gorm.DB, page models.Page) ([]BookingAgencyPayment, int64, error) {
	db = db.Model(BookingAgencyPayment{})
	list := []BookingAgencyPayment{}
	total := int64(0)
	status := constants.STATUS_ENABLE
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
