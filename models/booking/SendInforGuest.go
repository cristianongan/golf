package model_booking

import (
	"start/constants"
	"start/models"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// SendInforGuest
type SendInforGuest struct {
	models.ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"`  // Hang Golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(256);index"`   // San Golf
	BookingCode    string `json:"booking_code" gorm:"type:varchar(100);index"` // booking code
	BookingDate    string `json:"booking_date" gorm:"type:varchar(30);index"`  // Ex: 06/11/2022
	BookingName    string `json:"booking_name" gorm:"type:varchar(256);index"`
	GuestStyle     string `json:"guest_style" gorm:"type:varchar(256);index"`
	GuestStyleName string `json:"guest_style_name" gorm:"type:varchar(256)"`
	NumberPeople   int    `json:"number_people"`
	SendMethod     string `json:"send_method" gorm:"type:varchar(50)"`
	PhoneNumber    string `json:"phone_number" gorm:"type:varchar(50)"`
	Email          string `json:"email" gorm:"type:varchar(256)"`
	CmsUser        string `json:"cms_user" gorm:"type:varchar(100)"` // Cms User
}

// ======= CRUD ===========
func (item *SendInforGuest) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *SendInforGuest) Update(db *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *SendInforGuest) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *SendInforGuest) FindList(database *gorm.DB, page models.Page) ([]SendInforGuest, int64, error) {
	db := database.Model(SendInforGuest{})
	list := []SendInforGuest{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}
	if item.BookingCode != "" {
		db = db.Where(`booking_code COLLATE utf8mb4_general_ci LIKE ? OR booking_name COLLATE utf8mb4_general_ci LIKE ? 
		OR phone_number COLLATE utf8mb4_general_ci LIKE ? OR email COLLATE utf8mb4_general_ci LIKE ?`,
			"%"+item.BookingCode+"%", "%"+item.BookingCode+"%", "%"+item.BookingCode+"%", "%"+item.BookingCode+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *SendInforGuest) Delete(db *gorm.DB) error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
