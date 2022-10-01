package model_payment

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Single Payment
type SinglePayment struct {
	models.Model
	PartnerUid  string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string         `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	BookingUid  string         `json:"booking_uid" gorm:"type:varchar(100);index"` // Booking uid
	Bag         string         `json:"bag" gorm:"type:varchar(50)"`                // Golf bag
	BillCode    string         `json:"bill_code" gorm:"type:varchar(100);index"`
	BookingDate string         `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	PaymentDate string         `json:"payment_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	BagInfo     PaymentBagInfo `json:"bag_info,omitempty" gorm:"type:json"`

	Invoice            string `json:"invoice" gorm:"type:varchar(100)"`             // Invoice
	PaymentStatus      string `json:"payment_status" gorm:"type:varchar(50);index"` // PAID, UN_PAID, PARTIAL_PAID, DEBT
	PaymentType        string `json:"payment_type" gorm:"type:varchar(50);index"`   // CASH, VISA
	PrepaidFromBooking int64  `json:"prepaid_from_booking"`                         // Thanh toán trước từ khi booking (nếu có)
	Cashiers           string `json:"cashiers" gorm:"type:varchar(100);index"`      // Thu ngân, lấy từ acc cms
	TotalPaid          int64  `json:"total_paid" gorm:"type:varchar(100);index"`    // Số tiền thanh toán
	Note               string `json:"note" gorm:"type:varchar(200)"`                // Note
}

type PaymentBagInfo struct {
	CustomerName   string                       `json:"customer_name"`    // Tên khách hàng
	GuestStyle     string                       `json:"guest_style"`      // Guest Style
	GuestStyleName string                       `json:"guest_style_name"` // Guest Style Name
	CheckInTime    int64                        `json:"check_in_time"`    // Time Check In
	CheckOutTime   int64                        `json:"check_out_time"`   // Time Check Out
	MushPayInfo    model_booking.BookingMushPay `json:"mush_pay_info" `   // Mush Pay info
	SubBags        utils.ListSubBag             `json:"sub_bags"`         // List Sub Bags
}

func (item *PaymentBagInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item PaymentBagInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *SinglePayment) Create(db *gorm.DB) error {
	now := time.Now()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *SinglePayment) Update(mydb *gorm.DB) error {
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *SinglePayment) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *SinglePayment) Count(db *gorm.DB) (int64, error) {
	db = db.Model(SinglePayment{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *SinglePayment) FindList(db *gorm.DB, page models.Page, playerName string) ([]SinglePayment, int64, error) {
	db = db.Model(SinglePayment{})
	list := []SinglePayment{}
	total := int64(0)
	status := item.Model.Status
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

	if item.Bag != "" {
		db = db.Where("bag = ?", item.Bag)
	}

	if item.PaymentDate != "" {
		db = db.Where("payment_date = ?", item.PaymentDate)
	}

	if item.PaymentDate != "" {
		db = db.Where("booking_date = ?", item.PaymentDate)
	}

	if item.PaymentStatus != "" {
		db = db.Where("payment_status = ?", item.PaymentStatus)
	}

	if playerName != "" {
		db = db.Where("bag_info->'$.customer_name' LIKE ?", "%"+playerName+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

// func (item *Payment) Delete() error {
// 	if item.Model.Uid == "" {
// 		return errors.New("Primary key is undefined!")
// 	}
// 	return datasources.GetDatabase().Delete(item).Error
// }
