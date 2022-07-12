package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type TeeTimeSettings struct {
	ModelId
	PartnerUid    string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid     string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	TeeTime       string `json:"tee_time" gorm:"type:varchar(100);index"`
	TeeTimeStatus string `json:"tee_time_status" gorm:"type:varchar(100)"` // Trạng thái Tee Time: LOCKED, UNLOCK, DELETED
	Note          string `json:"note"`
}

func (item *TeeTimeSettings) IsDuplicated() bool {
	errFind := item.FindFirst()
	if errFind == nil {
		return true
	}
	return false
}

func (item *TeeTimeSettings) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *TeeTimeSettings) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *TeeTimeSettings) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *TeeTimeSettings) Count() (int64, error) {
	db := datasources.GetDatabase().Model(TeeTimeSettings{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *TeeTimeSettings) FindList(page Page) ([]TeeTimeSettings, int64, error) {
	db := datasources.GetDatabase().Model(TeeTimeSettings{})
	list := []TeeTimeSettings{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
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

func (item *TeeTimeSettings) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
