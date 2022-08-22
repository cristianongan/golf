package models

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type LockTurn struct {
	ModelId
	PartnerUid    string  `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid     string  `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	BookingDate   string  `json:"booking_date" gorm:"type:varchar(100)"`      // Ngày Booking
	TeeTimeLock   ListTee `json:"tee_time_lock,omitempty" gorm:"type:json"`   // Danh sách các teetime sẽ lock
	TeeTimeStatus string  `json:"tee_time_status" gorm:"type:varchar(100)"`   // Trạng thái Tee Time: LOCKED, UNLOCK, DELETED
	Note          string  `json:"note"`
}

type TeeInfo struct {
	TeeTime string `json:"tee_time" gorm:"type:varchar(100)"`
	TeeType string `json:"tee_type" gorm:"type:varchar(100)"`
}

type ListTee []TeeInfo

func (item *ListTee) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListTee) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *LockTurn) IsDuplicated() bool {
	errFind := item.FindFirst()
	return errFind == nil
}

func (item *LockTurn) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *LockTurn) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *LockTurn) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *LockTurn) Count() (int64, error) {
	db := datasources.GetDatabase().Model(LockTurn{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *LockTurn) FindList(page Page) ([]LockTurn, int64, error) {
	db := datasources.GetDatabase().Model(LockTurn{})
	list := []LockTurn{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
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

func (item *LockTurn) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
