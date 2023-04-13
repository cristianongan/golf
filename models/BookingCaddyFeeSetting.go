package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// CaddieFee setting
type BookingCaddyFeeSetting struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Fee        int64  `json:"fee"`                                        // phí tương ứng
	FromDate   int64  `json:"from_date" gorm:"index"`                     // Áp dụng từ ngày
	ToDate     int64  `json:"to_date" gorm:"index"`                       // Áp dụng tới ngày
	Name       string `json:"name" gorm:"type:varchar(256)"`
}

type BookingCaddyFeeSettingRes struct {
	Fee  int64  `json:"fee"`
	Name string `json:"name"`
}

func (item *BookingCaddyFeeSetting) IsValidated() bool {
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.Fee == 0 {
		return false
	}
	return true
}

func (item *BookingCaddyFeeSetting) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *BookingCaddyFeeSetting) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *BookingCaddyFeeSetting) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *BookingCaddyFeeSetting) Count(database *gorm.DB) (int64, error) {
	db := database.Model(BookingCaddyFeeSetting{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *BookingCaddyFeeSetting) FindAll(database *gorm.DB) ([]BookingCaddyFeeSetting, error) {
	db := database.Model(BookingCaddyFeeSetting{})
	list := []BookingCaddyFeeSetting{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db = db.Where("status = ENABLE")
	db = db.Order("created_at desc")
	db.Find(&list)

	return list, db.Error
}

func (item *BookingCaddyFeeSetting) FindBookingCaddieFee(database *gorm.DB) ([]BookingCaddyFeeSetting, error) {
	db := database.Model(BookingCaddyFeeSetting{})
	list := []BookingCaddyFeeSetting{}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	db = db.Where("status = ?", constants.STATUS_ENABLE)
	db = db.Order("created_at desc")

	db.Limit(1).Find(&list)
	return list, db.Error
}

func (item *BookingCaddyFeeSetting) FindList(database *gorm.DB, page Page, isGetAll bool) ([]BookingCaddyFeeSetting, int64, error) {
	db := database.Model(BookingCaddyFeeSetting{})
	list := []BookingCaddyFeeSetting{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.Status != "" {
		db = db.Where("status in (?)", strings.Split(item.Status, ","))
	}
	db.Count(&total)

	if isGetAll {
		if total > 0 && int64(page.Offset()) < total {
			db = page.Setup(db).Find(&list)
		}
	} else {
		db = db.Find(&list)
	}
	return list, total, db.Error
}

func (item *BookingCaddyFeeSetting) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
