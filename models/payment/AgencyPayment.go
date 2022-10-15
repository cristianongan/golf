package model_payment

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/models"
	model_booking "start/models/booking"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Single Payment
type AgencyPayment struct {
	models.Model
	PartnerUid  string            `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid   string            `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingCode string            `json:"booking_code" gorm:"type:varchar(100);index"` // Booking code
	BookingDate string            `json:"booking_date" gorm:"type:varchar(30);index"`  // Ex: 06/11/2022
	PaymentDate string            `json:"payment_date" gorm:"type:varchar(30);index"`  // Ex: 06/11/2022
	AgencyInfo  PaymentAgencyInfo `json:"agency_info,omitempty" gorm:"type:json"`
	AgencyId    int64             `json:"agency_id" gorm:"index"` // agency id

	PlayerBook   string `json:"player_book" gorm:"type:varchar(100)"` // Player book
	NumberPeople int    `json:"number_people"`                        // Sốn người chơi

	Invoice          string `json:"invoice" gorm:"type:varchar(100)"`                  // Invoice
	Cashiers         string `json:"cashiers" gorm:"type:varchar(100);index"`           // Thu ngân, lấy từ acc cms
	PaymentForPlayer string `json:"payment_for_player" gorm:"type:varchar(100);index"` // Thanh toán cho player
	Note             string `json:"note" gorm:"type:varchar(200)"`                     // Note

	TotalPaid           int64 `json:"total_paid"`             // Tổng số tiền thanh toán, Bao gồm tiền của đại lý đồng ý thanh toán cho khách
	TotalAmount         int64 `json:"total_amount"`           // Tổng chi phí phải thanh toán cho sân
	TotalFeeFromBooking int64 `json:"total_fee_from_booking"` // Tổng số tiền booking từ app trả về/ 1 booking Code.
	PaymentAgencyAmount int64 `json:"payment_agency_amount"`  // Ghi nhận số tiền đại lý thanh toán cho golfer nếu golfer thuộc đại lý
	PrepaidFromBooking  int64 `json:"prepaid_from_booking"`   // Số tiền thanh toán trên app hoặc booking tại sân
}

type PaymentAgencyInfo struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`             // Tên khách hàng
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name
}

func (item *PaymentAgencyInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item PaymentAgencyInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// Update Total Amount
func (item *AgencyPayment) UpdateTotalAmount(db *gorm.DB, isUdp bool) {
	booksR := model_booking.Booking{
		PartnerUid:  item.PartnerUid,
		CourseUid:   item.CourseUid,
		BookingCode: item.BookingCode,
	}

	listBook, errF := booksR.FindAllBookingOTA(db)

	totalAmount := int64(0)

	if errF == nil {
		for _, v := range listBook {
			totalAmount += v.MushPayInfo.MushPay
		}
	} else {
		log.Println("AgencyPayment UpdateTotalAmount errF", errF.Error())
	}

	item.TotalAmount = totalAmount
	item.PaymentAgencyAmount = totalAmount

	if isUdp {
		errUdp := item.Update(db)
		if errUdp != nil {
			log.Println("AgencyPayment UpdateTotalAmount errUdp", errUdp.Error())
		}
	}
}

// Update Total Paid
func (item *AgencyPayment) UpdateTotalPaid(db *gorm.DB) {
	if item.Uid == "" {
		return
	}
	//Total agency paid
	agencyPaymentItem := AgencyPaymentItem{
		PaymentUid: item.Uid,
	}
	listItem, err := agencyPaymentItem.FindAll(db)
	if err != nil {
		return
	}

	totalPaid := item.PrepaidFromBooking
	for _, v := range listItem {
		totalPaid += v.Paid
	}

	item.TotalPaid = totalPaid

	errUdp := item.Update(db)
	if errUdp != nil {
		log.Println("UpdateTotalPaid errUdp", errUdp.Error())
	}
}

func (item *AgencyPayment) Create(db *gorm.DB) error {
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

func (item *AgencyPayment) Update(mydb *gorm.DB) error {
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AgencyPayment) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *AgencyPayment) Count(db *gorm.DB) (int64, error) {
	db = db.Model(AgencyPayment{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *AgencyPayment) FindList(db *gorm.DB, page models.Page) ([]AgencyPayment, int64, error) {
	db = db.Model(AgencyPayment{})
	list := []AgencyPayment{}
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

	// if item.PaymentDate != "" {
	// 	db = db.Where("payment_date = ?", item.PaymentDate)
	// }

	if item.PaymentDate != "" {
		db = db.Where("booking_date = ?", item.PaymentDate)
	}

	// if item.PaymentStatus != "" {
	// 	db = db.Where("payment_status = ?", item.PaymentStatus)
	// }

	// if playerName != "" {
	// 	db = db.Where("bag_info->'$.customer_name' LIKE ?", "%"+playerName+"%")
	// }

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
