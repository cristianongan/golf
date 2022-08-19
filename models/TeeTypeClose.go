package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type TeeTypeClose struct {
	ModelId
	PartnerUid       string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid        string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	BookingSettingId int64  `json:"booking_setting_id" gorm:"type:varchar(50)"`
	DateTime         string `json:"date_time" gorm:"type:varchar(100)"`
	Note             string `json:"note"`
}

func (item *TeeTypeClose) IsDuplicated() bool {
	errFind := item.FindFirst()
	return errFind == nil
}

func (item *TeeTypeClose) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *TeeTypeClose) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *TeeTypeClose) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *TeeTypeClose) Count() (int64, error) {
	db := datasources.GetDatabase().Model(TeeTypeClose{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *TeeTypeClose) FindList(page Page) ([]TeeTypeClose, int64, error) {
	db := datasources.GetDatabase().Model(TeeTypeClose{})
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

func (item *TeeTypeClose) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
