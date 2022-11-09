package model_payment

import (
	"start/constants"
	"start/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Booking Agency Payment
type BookingAgencyPayment struct {
	models.ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingCode string `json:"booking_code" gorm:"type:varchar(100);index"` // Booking code
	AgencyId    int64  `json:"agency_id" gorm:"index"`                      // agency id
	BookingUid  string `json:"booking_uid" gorm:"type:varchar(100);index"`  // Booking Uid
	FeeType     string `json:"fee_type" gorm:"type:varchar(50);index"`      // Fee Type: GOLF_FEE, BUGGY_FEE, CADDIE_FEE
	Name        string `json:"name" gorm:"type:varchar(100)"`               // Ex: Buggy (1/2)xe
	Fee         int64  `json:"fee"`
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
