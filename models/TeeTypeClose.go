package models

import (
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TeeTypeClose struct {
	ModelId
	PartnerUid       string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid        string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	BookingSettingId int64  `json:"booking_setting_id" gorm:"type:varchar(50)"`
	DateTime         string `json:"date_time" gorm:"type:varchar(100)"`
	Note             string `json:"note"`
}

func (item *TeeTypeClose) IsDuplicated(db *gorm.DB) bool {
	errFind := item.FindFirst(db)
	return errFind == nil
}

func (item *TeeTypeClose) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}
	return db.Create(item).Error
}

func (item *TeeTypeClose) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *TeeTypeClose) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *TeeTypeClose) Count(database *gorm.DB) (int64, error) {
	db := database.Model(TeeTypeClose{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *TeeTypeClose) FindList(database *gorm.DB, page Page) ([]TeeTypeClose, int64, error) {
	db := database.Model(TeeTypeClose{})
	list := []TeeTypeClose{}
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
	if item.CreatedAt != 0 {
		db = db.Where("date_time = ?", item.DateTime)
	}
	if item.BookingSettingId != 0 {
		db = db.Where("booking_setting_id = ?", item.BookingSettingId)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *TeeTypeClose) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
