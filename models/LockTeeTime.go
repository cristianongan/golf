package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type LockTeeTime struct {
	ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	TeeTime        string `json:"tee_time" gorm:"type:varchar(100)"`
	CurrentTeeTime string `json:"current_tee_time" gorm:"type:varchar(100);index"`
	TeeTimeStatus  string `json:"tee_time_status" gorm:"type:varchar(100)"` // Trạng thái Tee Time: LOCKED, UNLOCK, DELETED
	DateTime       string `json:"date_time" gorm:"type:varchar(100)"`       // Ngày mà user update Tee Time
	TeeType        string `json:"tee_type" gorm:"type:varchar(100)"`        // TeeType: 1,10,1A ...
	Note           string `json:"note"`
}

func (item *LockTeeTime) IsDuplicated() bool {
	errFind := item.FindFirst()
	if errFind == nil {
		return true
	}
	return false
}

func (item *LockTeeTime) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *LockTeeTime) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *LockTeeTime) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *LockTeeTime) Count() (int64, error) {
	db := datasources.GetDatabase().Model(LockTeeTime{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *LockTeeTime) FindList(page *Page) ([]LockTeeTime, int64, error) {
	db := datasources.GetDatabase().Model(LockTeeTime{})
	list := []LockTeeTime{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.CreatedAt != 0 {
		db = db.Where("date_time = ?", item.DateTime)
	}
	if item.CurrentTeeTime != "" {
		db = db.Where("current_tee_time = ?", item.CurrentTeeTime)
	}

	db.Count(&total)

	if page != nil {
		if total > 0 && int64(page.Offset()) < total {
			db = page.Setup(db).Find(&list)
		}
	} else {
		db = db.Find(&list)
	}
	return list, total, db.Error
}

func (item *LockTeeTime) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
