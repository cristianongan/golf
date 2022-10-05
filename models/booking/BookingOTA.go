package model_booking

import (
	"start/constants"
	"start/models"
	"time"

	"gorm.io/gorm"
)

type BookingOta struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	PlayerName string `json:"player_name" gorm:"type:varchar(100)"`       //
	Contact    string `json:"contact" gorm:"type:varchar(100);index"`     // điện thoại;email

	Note         string `json:"note" gorm:"type:varchar(300)"`       // Note
	NumBook      int    `json:"num_book"`                            // số lượng book (<=4) và phụ thuộc số lượng chỗ trống còn lại của teetime
	Holes        int    `json:"holes"`                               // Số hố khi booking
	IsMainCourse bool   `json:"is_main_course"`                      // (true: book vào sân A. False: sân B
	Tee          string `json:"tee" gorm:"type:varchar(50)"`         //
	DateStr      string `json:"date_str" gorm:"type:varchar(50)"`    // Date
	TeeOffStr    string `json:"tee_off_str" gorm:"type:varchar(50)"` // Tee Off Str

	AgentCode   string `json:"agent_code" gorm:"type:varchar(100);index"`   // agent code
	GuestStyle  string `json:"guest_style" gorm:"type:varchar(200);index"`  // Guest Style
	BookingCode string `json:"booking_code" gorm:"type:varchar(256);index"` // Guest Style Name

	CaddieFee int64 `json:"caddie_fee"`
	BuggyFee  int64 `json:"buggy_fee"`
	GreenFee  int64 `json:"green_fee"`

	EmailConfirm string `json:"email_confirm" gorm:"type:varchar(100)"` // Mail verify
}

func (item *BookingOta) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingOta) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingOta) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingOta) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BookingOta{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}
