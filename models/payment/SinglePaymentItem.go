package model_payment

import (
	"start/constants"
	"start/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Single Payment item
type SinglePaymentItem struct {
	models.Model
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	BookingUid  string `json:"booking_uid" gorm:"type:varchar(100);index"` // Booking uid
	BillCode    string `json:"bill_code" gorm:"type:varchar(100);index"`   // Bill Code
	Bag         string `json:"bag" gorm:"type:varchar(50)"`                // Golf bag
	PaymentUid  string `json:"payment_uid" gorm:"type:varchar(100);index"`
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	PaymentType string `json:"payment_type" gorm:"type:varchar(50);index"` // CASH, VISA
	Cashiers    string `json:"cashiers" gorm:"type:varchar(100);index"`    // Thu ngân, lấy từ acc cms
	Paid        int64  `json:"paid" gorm:"type:varchar(100)"`              // Số tiền thanh toán
	Note        string `json:"note" gorm:"type:varchar(200)"`              // Note
	BankType    string `json:"bank_type" gorm:"type:varchar(20)"`          // Chuyển khoản qua VCB,BIDV...
}

type BookingSinglePayment struct {
	SinglePaymentItem
	CustomerBookingName string `json:"customer_booking_name"`
	CustomerName        string `json:"customer_name"`
	Amount              int64  `json:"amount"`
}

func (item *SinglePaymentItem) Create(db *gorm.DB) error {
	now := time.Now()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	uid := uuid.New()
	item.Model.Uid = uid.String()

	return db.Create(item).Error
}

func (item *SinglePaymentItem) Update(mydb *gorm.DB) error {
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *SinglePaymentItem) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *SinglePaymentItem) Count(db *gorm.DB) (int64, error) {
	db = db.Model(SinglePaymentItem{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *SinglePaymentItem) FindAll(db *gorm.DB) ([]SinglePaymentItem, error) {
	db = db.Model(SinglePaymentItem{})
	list := []SinglePaymentItem{}
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

	if item.PaymentUid != "" {
		db = db.Where("payment_uid = ?", item.PaymentUid)
	}

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	if item.Bag != "" {
		db = db.Where("bag = ?", item.Bag)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *SinglePaymentItem) FindList(db *gorm.DB, page models.Page) ([]SinglePaymentItem, int64, error) {
	db = db.Model(SinglePaymentItem{})
	list := []SinglePaymentItem{}
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

	if item.BillCode != "" {
		db = db.Where("bill_code = ?", item.BillCode)
	}

	if item.BookingUid != "" {
		db = db.Where("booking_uid = ?", item.BookingUid)
	}

	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	if item.Bag != "" {
		db = db.Where("bag = ?", item.Bag)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *SinglePaymentItem) FindAllTransfer(db *gorm.DB) ([]BookingSinglePayment, error) {
	db = db.Model(SinglePaymentItem{})
	list := []BookingSinglePayment{}
	db = db.Select("single_payment_items.*, bookings.customer_name, bookings.customer_booking_name, SUM(CAST(single_payment_items.paid as SIGNED)) as amount")
	db = db.Joins("LEFT JOIN bookings ON bookings.uid = single_payment_items.booking_uid")
	if item.PartnerUid != "" {
		db = db.Where("single_payment_items.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("single_payment_items.course_uid = ?", item.CourseUid)
	}

	if item.BookingDate != "" {
		db = db.Where("single_payment_items.booking_date = ?", item.BookingDate)
	}

	db = db.Where(`single_payment_items.payment_type = 'TRANSFER'`)
	db = db.Where(`single_payment_items.status <> 'DELETE'`)
	db.Group("single_payment_items.bag")
	db.Group("single_payment_items.payment_type")
	db.Find(&list)
	return list, db.Error
}

// func (item *Payment) Delete() error {
// 	if item.Model.Uid == "" {
// 		return errors.New("Primary key is undefined!")
// 	}
// 	return datasources.GetDatabase().Delete(item).Error
// }
