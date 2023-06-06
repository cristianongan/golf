package models_agency_booking

import (
	"errors"
	"start/models"
	"time"

	"gorm.io/gorm"
)

type AgencyBookingInfo struct {
	models.ModelId
	TransactionId string `json:"transaction_id"`                 // mã giao dịch
	BookingDate   string `json:"booking_date"`                   // dd/mm/yyyy
	CmsUser       string `json:"cms_user"`                       // Acc Operator Tạo (Bỏ lấy theo token)
	PartnerUid    string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid     string `json:"course_uid" binding:"required"`  // San Golf
	CourseType    string `json:"course_type"`
	HoleBooking   int    `json:"hole_booking"` // Số hố
	Hole          int    `json:"hole"`         // Số hố check
	TeeType       string `json:"tee_type"`     // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeePath       string `json:"tee_path"`     // MORNING, NOON, NIGHT (k required cái này vì có case checking k qua booking)
	TurnTime      string `json:"turn_time"`    // Ex: 16:26 (k required cái này vì có case checking k qua booking)
	TeeTime       string `json:"tee_time"`     // Ex: 16:26 Tee time là thời gian tee off dự kiến (k required cái này vì có case checking k qua booking)
	RowIndex      *int   `json:"row_index"`    // index trong Flight

	// Guest booking
	GuestStyle           string  `json:"guest_style"`            // Guest Style
	GuestStyleName       string  `json:"guest_style_name"`       // Guest Style Name
	CustomerName         string  `json:"customer_name"`          // Tên khách hàng
	CustomerBookingEmail *string `json:"customer_booking_email"` // Email khách hàng
	CustomerBookingName  string  `json:"customer_booking_name"`  // Tên khách hàng đặt booking
	CustomerBookingPhone string  `json:"customer_booking_phone"` // SDT khách hàng đặt booking
	CustomerIdentify     string  `json:"customer_identify"`      // passport/cccd
	Nationality          string  `json:"nationality"`            // Nationality
}

func (_ *AgencyBookingInfo) CreateBatch(list []AgencyBookingInfo, db *gorm.DB) error {
	db.Model(&AgencyBookingInfo{})

	now := time.Now().Unix()

	for _, item := range list {
		item.CreatedAt = now
		item.UpdatedAt = now
	}

	return db.CreateInBatches(list, 100).Error
}

func (_ *AgencyBookingInfo) FindListByTransactionId(transactionId string, db *gorm.DB) ([]AgencyBookingInfo, error) {
	db.Model(&AgencyBookingInfo{})

	list := []AgencyBookingInfo{}
	var err error

	if transactionId == "" {
		return list, errors.New("transaction id is required")
	}

	db = db.Where(" transaction_id = (?) ", transactionId)

	err = db.Find(&list).Error

	return list, err
}

func (item *AgencyBookingInfo) DeleteBatch(db *gorm.DB) error {
	return db.Delete(item).Error
}
