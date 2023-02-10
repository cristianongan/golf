package model_payment

import (
	"start/constants"
	"start/models"
	"start/utils"
	"strings"

	"gorm.io/gorm"
)

// Booking Agency Payment
type BookingAgencyPayment struct {
	models.ModelId
	PartnerUid  string                               `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid   string                               `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingCode string                               `json:"booking_code" gorm:"type:varchar(100);index"` // Booking code
	AgencyId    int64                                `json:"agency_id" gorm:"index"`                      // agency id
	BookingUid  string                               `json:"booking_uid" gorm:"type:varchar(100);index"`  // Booking Uid
	CaddieId    string                               `json:"caddie_id" gorm:"type:varchar(100)"`          // Caddie Id
	FeeData     utils.ListBookingAgencyPayForBagData `json:"fee_data,omitempty" gorm:"type:json"`         // fee data
}

func (item *BookingAgencyPayment) GetTotalFee() int64 {
	if item.FeeData == nil || len(item.FeeData) == 0 {
		return 0
	}

	totalFee := int64(0)

	for _, v := range item.FeeData {
		totalFee += v.Fee
	}

	return totalFee
}

func (item *BookingAgencyPayment) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingAgencyPayment) Update(mydb *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
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
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.AgencyId > 0 {
		db = db.Where("agency_id = ?", item.AgencyId)
	}

	if item.BookingCode != "" {
		db = db.Where("booking_code = ?", item.BookingCode)
	}

	if item.BookingUid != "" {
		db = db.Where("booking_uid = ?", item.BookingUid)
	}

	if item.AgencyId > 0 {
		db = db.Where("agency_id = ?", item.AgencyId)
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
