package models

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/utils"
	"strings"

	"gorm.io/gorm"
)

// Hãng Golf
type Notification struct {
	ModelId
	PartnerUid         string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid          string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Type               string `json:"type" gorm:"type:varchar(100)"`              // Loại noti
	Title              string `json:"title" gorm:"type:varchar(256)"`
	NotificationStatus string `json:"noti_status" gorm:"type:varchar(50)"`  // Trạng thái của noti
	UserCreate         string `json:"user_create" gorm:"type:varchar(100)"` // Người tạo noti
	UserApprove        string `json:"user_update" gorm:"type:varchar(100)"` // Người duyệt noti
	IsRead             *bool  `json:"is_read" gorm:"default:0"`             // Trạng thái đã xem của noti
	Note               string `json:"note" gorm:"type:varchar(500)"`
	Content            []byte `json:"content" gorm:"type:json"`
	DateApproved       int64  `json:"date_approved"`
	Role               string `json:"role"`
}

type NotificationResponse struct {
}

type CaddieContentNoti struct {
	Id           int64  `json:"id"`   // id của object tạo noti
	Code         string `json:"code"` //  caddie code
	Type         string `json:"type"`
	NumberDayOff int    `json:"number_day_off"`
	FromDay      string `json:"from_day"`
	ToDay        string `json:"to_day"`
	Hour         string `json:"hour"`
}

type CaddieWCINoti struct {
	Caddies   []string `json:"caddies"`    // danh sách caddie
	ApplyDate string   `json:"apply_date"` //  ngày áp dụng
}

func (item *CaddieContentNoti) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item CaddieContentNoti) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ======= CRUD ===========
func (item *Notification) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *Notification) Update(database *gorm.DB) error {
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := database.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Notification) FindFirst(database *gorm.DB) error {
	db := database.Model(Notification{})
	return db.Where(item).First(item).Error
}

func (item *Notification) Count(database *gorm.DB) (int64, error) {
	db := database.Model(Notification{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Notification) FindList(database *gorm.DB, page Page) ([]Notification, int64, error) {
	db := database.Model(Notification{})
	list := []Notification{}
	total := int64(0)
	status := item.Status
	item.Status = ""
	db = db.Where(item)

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Notification) Delete(database *gorm.DB) error {
	return database.Delete(item).Error
}
