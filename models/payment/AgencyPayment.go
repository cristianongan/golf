package model_payment

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
	"strings"

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

	TotalPaid   int64 `json:"total_paid"`   // D Tổng số tiền thanh toán, Bao gồm tiền của đại lý đồng ý thanh toán cho khách
	TotalAmount int64 `json:"total_amount"` // A Tổng chi phí phải thanh toán cho sân
	// TotalFeeFromBooking int64 `json:"total_fee_from_booking"` // Tổng số tiền booking từ app trả về/ 1 booking Code.
	AgencyPaid int64 `json:"agency_paid"` // D Ghi nhận số tiền đại lý thanh toán cho golfer nếu golfer thuộc đại lý
	// PrepaidFromBooking  int64 `json:"prepaid_from_booking"`   // Số tiền thanh toán trên app hoặc booking tại sân
}

type PaymentAgencyInfo struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`             // Tên khách hàng
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name
}

type AgencyPaidForBagDetail struct {
	BookingAgencyPayment
	ListServiceItems []model_booking.BookingServiceItem `json:"list_service_items,omitempty"` // Guest Style
}

func (item *PaymentAgencyInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item PaymentAgencyInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

/*
Udp số người chơi và info user đặt book
*/
func (item *AgencyPayment) UpdatePlayBookInfo(db *gorm.DB, booking model_booking.Booking) {
	if booking.BookingCode == "" {
		return
	}
	bookOTA := model_booking.BookingOta{
		PartnerUid:  booking.PartnerUid,
		CourseUid:   booking.CourseUid,
		BookingCode: booking.BookingCode,
	}
	errFindBO := bookOTA.FindFirst(db)
	if errFindBO == nil {
		feeBooking := int64(bookOTA.NumBook) * (bookOTA.CaddieFee + bookOTA.BuggyFee + bookOTA.GreenFee)
		item.AgencyPaid = feeBooking
		item.PlayerBook = bookOTA.PlayerName
		item.NumberPeople = bookOTA.NumBook
	} else {
		// Find booking waiting
		bookR := model_booking.Booking{
			PartnerUid:  booking.PartnerUid,
			CourseUid:   booking.CourseUid,
			BookingCode: booking.BookingCode,
		}
		listBook, err := bookR.FindAllBookingOTA(db)
		if err == nil {
			if len(listBook) > 0 {
				item.NumberPeople = len(listBook)
				item.PlayerBook = listBook[0].CustomerBookingName
			}
		}
	}
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

	// Get agency paid
	agencyPaid := BookingAgencyPayment{
		PartnerUid:  item.PartnerUid,
		BookingCode: item.BookingCode,
		AgencyId:    item.AgencyId,
	}

	listAgencyPaid, errAP := agencyPaid.FindAll(db)
	agencyPaidFee := int64(0)
	if errAP == nil {
		for _, v := range listAgencyPaid {
			agencyPaidFee += v.GetTotalFee()
		}
		item.AgencyPaid = agencyPaidFee
	}

	item.TotalAmount = totalAmount + agencyPaidFee
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

	// totalPaid := item.PrepaidFromBooking
	totalPaid := int64(0)
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

	if errC == nil {
		//Add vào redis để check
		redisKey := utils.GetRedisKeyAgencyPaymentCreated(item.PartnerUid, item.CourseUid, item.BookingCode)
		datasources.SetCache(redisKey, "1", 10) // expried 10s
	}

	return errC
}

func (item *AgencyPayment) Update(mydb *gorm.DB) error {
	item.Model.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *AgencyPayment) FindFirst(db *gorm.DB) error {
	errF := db.Where(item).First(item).Error

	if errF != nil {
		redisKey := utils.GetRedisKeyAgencyPaymentCreated(item.PartnerUid, item.CourseUid, item.BookingCode)
		strValue, redisErr := datasources.GetCache(redisKey)
		if redisErr == nil && strValue != "" {
			log.Println("[PAYMENT] agency redis", redisKey)
			return nil
		}
	}

	return errF
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
