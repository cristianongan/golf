package models

import (
	"start/constants"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TeeTimeList struct {
	ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	TeeTime     string `json:"tee_time" gorm:"type:varchar(20);index"`     // Giờ tee time
	TeeType     string `json:"tee_type" gorm:"type:varchar(20)"`           // Loại Tee 1,10
	CourseType  string `json:"course_type" gorm:"type:varchar(20)"`        // Sân A,B,C
	BookingDate string `json:"booking_date" gorm:"type:varchar(30);index"` // Thời gian
	SlotEmpty   int    `json:"slot_empty"`                                 // slot còn trống
	Part        int    `json:"part"`
}

type TeePart struct {
	TeeTime string
	Part    int
}

func (item *TeeTimeList) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}
	return db.Create(item).Error
}

func (item *TeeTimeList) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *TeeTimeList) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *TeeTimeList) Count(database *gorm.DB) (int64, error) {
	db := database.Model(TeeTimeList{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *TeeTimeList) FindList(database *gorm.DB, page Page) ([]TeeTimeList, int64, error) {
	db := database.Model(TeeTimeList{})
	list := []TeeTimeList{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *TeeTimeList) FindAllList(database *gorm.DB) ([]TeeTimeList, int64, error) {
	db := database.Model(TeeTimeList{})
	list := []TeeTimeList{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.BookingDate != "" {
		db = db.Where("booking_date = ?", item.BookingDate)
	}

	db.Count(&total)

	db.Find(&list)

	return list, total, db.Error
}

func (item *TeeTimeList) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
