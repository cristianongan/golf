package model_booking

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"start/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Booking
type Booking struct {
	models.Model
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf

	Bag            string `json:"bag" gorm:"type:varchar(100);index"`         // Golf Bag
	Hole           int    `json:"hole"`                                       // Số hố
	GuestStyle     string `json:"guest_style" gorm:"type:varchar(200);index"` // Guest Style
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(256)"`  // Guest Style Name

	CardId       int64  `json:"card_id" gorm:"index"`                        // Card ID
	CustomerName string `json:"customer_name" gorm:"type:varchar(256)"`      // Tên khách hàng
	CustomerUid  string `json:"customer_uid" gorm:"type:varchar(256);index"` // Uid khách hàng

	TeeType    string `json:"tee_type" gorm:"type:varchar(50);index"` // Tee1, Tee10, Tea1A, Tea1B, Tea1C,
	TurnTime   string `json:"turn_time" gorm:"type:varchar(30)"`      // Ex: 16:26
	TeeOffTime string `json:"tee_off_time" gorm:"type:varchar(30)"`   // Ex: 16:26 --- dự kiến: expected
	RowIndex   int    `json:"row_index"`                              // index trong Flight

	BookingFeeInfo BookingFee `json:"booking_fee_info" gorm:"type:varchar(500)"` // Thông tin phí

}

type BookingFee struct {
}

func (item *Booking) Create() error {
	uid := uuid.New()
	now := time.Now()
	item.Model.Uid = item.CourseUid + "-" + utils.HashCodeUuid(uid.String())
	item.Model.CreatedAt = now.Unix()
	item.Model.UpdatedAt = now.Unix()
	if item.Model.Status == "" {
		item.Model.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Booking) Update() error {
	mydb := datasources.GetDatabase()
	item.Model.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Booking) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Booking) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Booking) FindList(page models.Page) ([]Booking, int64, error) {
	db := datasources.GetDatabase().Model(Booking{})
	list := []Booking{}
	total := int64(0)
	status := item.Model.Status
	item.Model.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Booking) Delete() error {
	if item.Model.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
