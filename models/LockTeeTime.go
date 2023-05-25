package models

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
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

type LockTeeTimeWithSlot struct {
	ModelId
	PartnerUid     string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid      string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	TeeTime        string `json:"tee_time" gorm:"type:varchar(100)"`
	CurrentTeeTime string `json:"current_tee_time" gorm:"type:varchar(100);index"`
	TeeTimeStatus  string `json:"tee_time_status" gorm:"type:varchar(100)"` // Trạng thái Tee Time: LOCKED, UNLOCK, DELETED
	DateTime       string `json:"date_time" gorm:"type:varchar(100)"`       // Ngày mà user update Tee Time
	TeeType        string `json:"tee_type" gorm:"type:varchar(100)"`        // TeeType: 1,10,1A ...
	CurrentCourse  string `json:"course_current,omitempty"`
	Note           string `json:"note"`
	Slot           int    `json:"slot"`
	Type           string `json:"type"`
	LockLevel      string `json:"lock_level"`
}

func (item *LockTeeTimeWithSlot) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item LockTeeTimeWithSlot) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *LockTeeTime) IsDuplicated(db *gorm.DB) bool {
	errFind := item.FindFirst(db)
	if errFind == nil {
		return true
	}
	return false
}

func (item *LockTeeTime) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	return db.Create(item).Error
}

func (item *LockTeeTime) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *LockTeeTime) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *LockTeeTime) Count(database *gorm.DB) (int64, error) {
	db := database.Model(LockTeeTime{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *LockTeeTime) FindList(database *gorm.DB, requestType string) ([]LockTeeTime, int64, error) {
	db := database.Model(LockTeeTime{})
	list := []LockTeeTime{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	if item.DateTime != "" {
		db = db.Where("date_time = ?", item.DateTime)
	}
	if item.CurrentTeeTime != "" {
		db = db.Where("current_tee_time = ?", item.CurrentTeeTime)
	}
	if item.TeeTime != "" {
		db = db.Where("tee_time = ?", item.TeeTime)
	}
	if requestType == "TURN_TIME" {
		db = db.Where("current_tee_time <> tee_time")
	}
	if requestType == "TEE_TIME" {
		db = db.Where("current_tee_time = tee_time OR current_tee_time is NULL")
	}

	db.Count(&total)
	db = db.Find(&list)

	return list, total, db.Error
}

func (item *LockTeeTime) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
