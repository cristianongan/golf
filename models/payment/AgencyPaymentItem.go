package model_payment

import (
	"start/constants"
	"start/models"
	"start/utils"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Single Payment item
type AgencyPaymentItem struct {
	models.Model
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingCode string `json:"booking_code" gorm:"type:varchar(100);index"` // Booking code
	PaymentUid  string `json:"payment_uid" gorm:"type:varchar(100);index"`
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	PaymentType string `json:"payment_type" gorm:"type:varchar(50);index"` // CASH, VISA
	Cashiers    string `json:"cashiers" gorm:"type:varchar(100);index"`    // Thu ngân, lấy từ acc cms
	Paid        int64  `json:"paid" gorm:"type:varchar(100)"`              // Số tiền thanh toán
	Note        string `json:"note" gorm:"type:varchar(200)"`              // Note
}

func (item *AgencyPaymentItem) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	uid := uuid.New()
	item.Model.Uid = uid.String()

	return db.Create(item).Error
}

func (item *AgencyPaymentItem) Update(mydb *gorm.DB) error {
	item.Model.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AgencyPaymentItem) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *AgencyPaymentItem) Count(db *gorm.DB) (int64, error) {
	db = db.Model(AgencyPaymentItem{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *AgencyPaymentItem) FindAll(db *gorm.DB) ([]AgencyPaymentItem, error) {
	db = db.Model(AgencyPaymentItem{})
	list := []AgencyPaymentItem{}
	status := constants.STATUS_ENABLE
	item.Model.Status = ""

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

	if item.PaymentUid != "" {
		db = db.Where("payment_uid = ?", item.PaymentUid)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *AgencyPaymentItem) FindList(db *gorm.DB, page models.Page) ([]AgencyPaymentItem, int64, error) {
	db = db.Model(AgencyPaymentItem{})
	list := []AgencyPaymentItem{}
	total := int64(0)
	status := constants.STATUS_ENABLE
	item.Model.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
