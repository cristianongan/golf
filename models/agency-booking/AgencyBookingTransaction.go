package models_agency_booking

import (
	"errors"
	"start/constants"
	"start/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgencyBookingTransaction struct {
	Id                   int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	CreatedAt            int64  `json:"created_at" gorm:"index"`
	CreatedBy            string `json:"created_by"`
	UpdatedAt            int64  `json:"updated_at"`
	TransactionId        string `json:"transaction_id" gorm:"index"`           // mã giao dịch ~ booking code
	CourseUid            string `json:"course_id" gorm:"" binding:"required"`  // mã sân
	PartnerUid           string `json:"partner_id" gorm:"" binding:"required"` //
	AgencyId             int64  `json:"agency_id"`                             // Mã đại lí
	PaymentStatus        string `json:"payment_status" gorm:""`                // trạng thái thanh toán
	TransactionStatus    string `json:"transaction_status" gorm:""`            // trạng thái đơn ~ trạng thái giao dịch
	BookingRequestStatus string `json:"booking_request_status" gorm:""`        // trạng thái yêu cầu booking với nhà cung cấp
	CustomerPhoneNumber  string `json:"customer_phone_number" gorm:""`         // điện thoại người đặt
	CustomerEmail        string `json:"customer_email" gorm:""`                // email người đặt
	CustomerName         string `json:"customer_name" gorm:""`                 // Tên người đặt
	BookingAmount        int64  `json:"booking_amount"`                        // giá book
	ServiceAmount        int64  `json:"service_amount"`                        // giá dịch vụ
	PaymentDueDate       int64  `json:"payment_due_date"`                      // ngày hết hạn thanh toán
	PlayerNote           string `json:"player_note" gorm:""`                   // ghi ch ú người chơi
	AgentNote            string `json:"agent_note" gorm:""`                    // ghi chú tổn đài viên
	PlayDate             int64  `json:"play_date"`                             // ngày chơi

	// thông tin hóa đơn
	PaymentType    string `json:"payment_type" gorm:"" binding:"required"` // loại thanh toán
	Company        string `json:"company" gorm:""`                         // công ty
	CompanyAddress string `json:"company_address" gorm:""`                 // địa chỉ công ty
	CompanyTax     string `json:"company_tax" gorm:""`                     // Mã số thuế
	ReceiptEmail   string `json:"receipt_email" gorm:""`                   // email nhận hóa đơn

}

func (item *AgencyBookingTransaction) Create(db *gorm.DB) error {
	now := time.Now().Unix()

	item.TransactionId = createTransactionId(item.AgencyId, item.CourseUid)
	item.CreatedAt = now
	item.UpdatedAt = now

	return db.Create(&item).Error
}

func (item *AgencyBookingTransaction) Update(db *gorm.DB) error {
	return db.Save(item).Error
}

func (item *AgencyBookingTransaction) updateTransactionStatus(oldData AgencyBookingTransaction, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// create history
		his := AgencyBookingTransactionHis{
			TransactionId:     oldData.TransactionId,
			CreatedAt:         oldData.CreatedAt,
			CreatedBy:         oldData.CreatedBy,
			TransactionStatus: oldData.TransactionStatus,
		}

		if err := his.Create(tx); err != nil {
			return err
		}

		// update
		return db.Save(item).Error
	})
}

func (item *AgencyBookingTransaction) FindList(from int64, to int64, page models.Page, db *gorm.DB) ([]AgencyBookingTransaction, int64, error) {
	db.Model(AgencyBookingTransaction{})
	var total int64
	list := []AgencyBookingTransaction{}

	db = db.Where(" updated_at = 0 ")

	paymentStatus := item.PaymentStatus
	transactionStatus := item.TransactionStatus
	bookingRequestStatus := item.BookingRequestStatus

	if paymentStatus != "" {
		db = db.Where("payment_status IN (?)", strings.Split(paymentStatus, ","))
	}

	if transactionStatus != "" {
		db = db.Where("transaction_status IN (?)", strings.Split(transactionStatus, ","))
	}

	if bookingRequestStatus != "" {
		db = db.Where("booking_request_status IN (?)", strings.Split(bookingRequestStatus, ","))
	}

	if item.PartnerUid != "" && item.PartnerUid != constants.ROOT_PARTNER_UID {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.TransactionId != "" {
		db = db.Where("transaction_id LIKE ?", "%"+item.TransactionId+"%")
	}

	if item.CustomerPhoneNumber != "" {
		db = db.Where("customer_phone_number LIKE ?", "%"+item.CustomerPhoneNumber+"%")
	}

	fromStr := strconv.FormatInt(from, 10)
	toStr := strconv.FormatInt(to, 10)

	db = db.Where("created_at between " + fromStr + " and " + toStr + " OR play_date between " + fromStr + " and " + toStr)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *AgencyBookingTransaction) FindHistories(page models.Page, db *gorm.DB) ([]AgencyBookingTransaction, int64, error) {
	list := []AgencyBookingTransaction{}
	var total int64

	if item.TransactionId == "" {
		return list, total, errors.New("transaction id is required")
	}

	db = db.Where(" transaction_id = ? ", item.TransactionId)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *AgencyBookingTransaction) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func createTransactionId(agencyId int64, courseUid string) string {
	rs := ""

	rs += strconv.FormatInt(agencyId, 10)

	if len(courseUid) > 2 {
		rs += courseUid[:2]
	} else {
		rs += courseUid
	}

	rs += strings.ReplaceAll(uuid.NewString(), "-", "")

	return strings.ToUpper(rs)
}
