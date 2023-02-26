package model_payment

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log"
	"start/constants"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Single Payment
type SinglePayment struct {
	models.Model
	PartnerUid  string         `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid   string         `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingUid  string         `json:"booking_uid" gorm:"type:varchar(100);index"`  // Booking uid
	BookingCode string         `json:"booking_code" gorm:"type:varchar(100);index"` // Booking code
	Bag         string         `json:"bag" gorm:"type:varchar(50);index"`           // Golf bag
	BillCode    string         `json:"bill_code" gorm:"type:varchar(100);index"`
	BookingDate string         `json:"booking_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	PaymentDate string         `json:"payment_date" gorm:"type:varchar(30);index"` // Ex: 06/11/2022
	BagInfo     PaymentBagInfo `json:"bag_info,omitempty" gorm:"type:json"`
	// TotalAmount int64          `json:"total_amount"` // Số tiền thanh toán

	Invoice       string `json:"invoice" gorm:"type:varchar(100)"`             // Invoice
	PaymentStatus string `json:"payment_status" gorm:"type:varchar(50);index"` // PAID, UN_PAID, PARTIAL_PAID, DEBT
	// PrepaidFromBooking int64  `json:"prepaid_from_booking"`                         // Thanh toán trước từ khi booking (nếu có)
	Cashiers   string `json:"cashiers" gorm:"type:varchar(100);index"` // Thu ngân, lấy từ acc cms
	TotalPaid  int64  `json:"total_paid"`                              // Số tiền thanh toán
	AgencyPaid int64  `json:"agency_paid"`                             // Số tiền đại lý thanh toán
	Note       string `json:"note" gorm:"type:varchar(200)"`           // Note
	Type       string `json:"type" gorm:"type:varchar(50);index"`      // SINGLE, AGENCY
	IsEInvoice int    `json:"is_einvoice"`                             // Đánh dấu lấy hóa đơn điện tử
}

type PaymentBagInfo struct {
	TeeTime        string                       `json:"tee_time"`         // Ex: 16:26 Tee time là thời gian tee off dự kiến
	Hole           int                          `json:"hole"`             // Số hố check in
	CustomerName   string                       `json:"customer_name"`    // Tên khách hàng
	GuestStyle     string                       `json:"guest_style"`      // Guest Style
	GuestStyleName string                       `json:"guest_style_name"` // Guest Style Name
	CheckInTime    int64                        `json:"check_in_time"`    // Time Check In
	CheckOutTime   int64                        `json:"check_out_time"`   // Time Check Out
	MushPayInfo    model_booking.BookingMushPay `json:"mush_pay_info" `   // Mush Pay info
	SubBags        utils.ListSubBag             `json:"sub_bags"`         // List Sub Bags
	MainBags       utils.ListSubBag             `json:"main_bags"`
}

type SinglePaymentWithDebt struct {
	SinglePayment
	CountDebt int `json:"count_debt"`
}

func (item *PaymentBagInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item PaymentBagInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

/*
Update payment status
*/
func (item *SinglePayment) UpdatePaymentStatus(bagStatus string, db *gorm.DB) {
	// find single payment items
	singlePaymentItem := SinglePaymentItem{}
	singlePaymentItem.PaymentUid = item.Uid
	listItem, errF := singlePaymentItem.FindAll(db)

	totalAmount := item.BagInfo.MushPayInfo.MushPay

	isDebt := true
	if errF != nil || len(listItem) == 0 {
		isDebt = false
	}

	for _, v := range listItem {
		if v.PaymentType != constants.PAYMENT_STATUS_DEBIT {
			isDebt = false
		}
	}

	if isDebt {
		item.PaymentStatus = constants.PAYMENT_STATUS_DEBIT
		return
	}

	if item.TotalPaid <= 0 {
		if bagStatus == constants.BAG_STATUS_CHECK_OUT || bagStatus == constants.BAG_STATUS_CANCEL {
			if totalAmount == 0 {
				//PAID
				item.PaymentStatus = constants.PAYMENT_STATUS_PAID
			} else {
				//UN_PAID
				item.PaymentStatus = constants.PAYMENT_STATUS_UN_PAID
			}
		} else {
			//UN_PAID
			item.PaymentStatus = constants.PAYMENT_STATUS_UN_PAID
		}
		return
	}

	if totalAmount > item.TotalPaid && item.TotalPaid > 0 {
		//PARTIAL_PAID
		item.PaymentStatus = constants.PAYMENT_STATUS_PARTIAL_PAID
		return
	}

	if totalAmount <= item.TotalPaid {
		//PAID
		item.PaymentStatus = constants.PAYMENT_STATUS_PAID
		return
	}
}

func (item *SinglePayment) UpdateTotalPaid(db *gorm.DB) {
	if item.Uid == "" {
		return
	}
	totalPaid := int64(0)
	//Total agency paid
	agencyPaymentItem := SinglePaymentItem{
		PaymentUid: item.Uid,
	}
	listItem, err := agencyPaymentItem.FindAll(db)
	if err != nil || len(listItem) == 0 {
		totalPaid = 0
	} else {
		for _, v := range listItem {
			totalPaid += v.Paid
		}
	}

	item.TotalPaid = totalPaid
	if item.BookingUid != "" {
		//booking
		booking := model_booking.Booking{}
		booking.Uid = item.BookingUid
		errFB := booking.FindFirst(db)
		if errFB == nil {
			item.UpdatePaymentStatus(booking.BagStatus, db)
		}
	}

	errUdp := item.Update(db)
	if errUdp != nil {
		log.Println("UpdateTotalPaid errUdp", errUdp.Error())
	}
}

func (item *SinglePayment) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	uid := uuid.New()
	item.Model.Uid = uid.String()

	item.Invoice = constants.CONS_INVOICE + "-" + utils.HashCodeUuid(item.Uid)

	errC := db.Create(item).Error

	return errC
}

func (item *SinglePayment) Update(mydb *gorm.DB) error {
	item.Model.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *SinglePayment) FindFirst(db *gorm.DB) error {
	errF := db.Where(item).First(item).Error
	return errF
}

func (item *SinglePayment) Count(db *gorm.DB) (int64, error) {
	db = db.Model(SinglePayment{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *SinglePayment) FindAllForAgency(db *gorm.DB) ([]SinglePayment, error) {
	db = db.Model(SinglePayment{})
	list := []SinglePayment{}

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

	if item.Type != "" {
		db = db.Where("type = ?", item.Type)
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

func (item *SinglePayment) FindListWithJoin(db *gorm.DB, page models.Page, playerName string) ([]SinglePaymentWithDebt, int64, error) {
	db = db.Model(SinglePayment{})
	list := []SinglePaymentWithDebt{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("single_payments.partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("single_payments.course_uid = ?", item.CourseUid)
	}

	if item.Bag != "" {
		db = db.Where("single_payments.bag = ?", item.Bag)
	}

	if item.Type != "" {
		db = db.Where("single_payments.type = ?", item.Type)
	}

	if item.PaymentDate != "" {
		db = db.Where("single_payments.payment_date = ?", item.PaymentDate)
	}

	if item.PaymentStatus != "" {
		db = db.Where("single_payments.payment_status = ?", item.PaymentStatus)
	}

	if playerName != "" {
		db = db.Where("single_payments.bag_info->'$.customer_name' LIKE ?", "%"+playerName+"%")
	}

	db.Joins("LEFT JOIN single_payment_items ON single_payment_items.payment_uid = single_payments.uid")
	db.Select("single_payments.*, SUM(single_payment_items.payment_type ='DEBIT') as count_debt")
	db.Group("single_payments.uid")
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *SinglePayment) Delete(db *gorm.DB) error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
