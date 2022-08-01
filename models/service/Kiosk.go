package model_service

import (
	"errors"
	"start/constants"
	"start/datasources"
	"start/models"
	"strings"
	"time"
)

// Kiosk
type Kiosk struct {
	models.ModelId
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	KioskName   string `json:"kiosk_name" gorm:"type:varchar(256)"`        // Tên
	ServiceType string `json:"service_type" gorm:"type:varchar(50)"`       // Loại rental, kiosk, proshop
	KioskType   string `json:"kiosk_type" gorm:"type:varchar(50)"`         // Kiểu Kiosk (Mini Bar, Mini Restaurant,...)
}

func (item *Kiosk) IsValidated() bool {
	if item.KioskName == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.KioskType == "" {
		return false
	}
	if item.ServiceType == "" {
		return false
	}
	return true
}

func (item *Kiosk) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Kiosk) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Kiosk) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Kiosk) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Kiosk{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Kiosk) FindList(page models.Page) ([]Kiosk, int64, error) {
	db := datasources.GetDatabase().Model(Kiosk{})
	list := []Kiosk{}
	total := int64(0)
	status := item.ModelId.Status

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.KioskName != "" {
		db = db.Where("name LIKE ?", "%"+item.KioskName+"%")
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Kiosk) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
