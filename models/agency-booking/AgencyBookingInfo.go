package models_agency_booking

import (
	"errors"
	"start/models"
	"time"

	"gorm.io/gorm"
)

type AgencyBookingInfo struct {
	models.ModelId
	TransactionId string `json:"transaction_id" gorm:"type:varchar(100);index"`           // mã giao dịch
	BookingDate   string `json:"booking_date" gorm:"type:varchar(100)"`                   // dd/mm/yyyy
	PartnerUid    string `json:"partner_uid" binding:"required" gorm:"type:varchar(100)"` // Hang Golf
	CourseUid     string `json:"course_uid" binding:"required" gorm:"type:varchar(256)"`  // San Golf
	CourseType    string `json:"course_type" gorm:"type:varchar(100)"`
	HoleBooking   int    `json:"hole_booking"`                      // Số hố
	Hole          int    `json:"hole"`                              // Số hố check
	TeeType       string `json:"tee_type" gorm:"type:varchar(50)"`  // 1, 1A, 1B, 1C, 10, 10A, 10B (k required cái này vì có case checking k qua booking)
	TeePath       string `json:"tee_path" gorm:"type:varchar(50)"`  // MORNING, NOON, NIGHT (k required cái này vì có case checking k qua booking)
	TurnTime      string `json:"turn_time" gorm:"type:varchar(30)"` // Ex: 16:26 (k required cái này vì có case checking k qua booking)
	TeeTime       string `json:"tee_time" gorm:"type:varchar(30)"`  // Ex: 16:26 Tee time là thời gian tee off dự kiến (k required cái này vì có case checking k qua booking)
	RowIndex      *int   `json:"row_index"`                         // index trong Flight
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
