package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Group Fee
type Holiday struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`              // Ten Holiday
	Note       string `json:"note" gorm:"type:varchar(256)"`
	From       string `json:"from" gorm:"type:varchar(100)"`
	To         string `json:"to" gorm:"type:varchar(100)"`
	Year       string `json:"year" gorm:"type:varchar(100)"`
}

type HolidayResponse struct {
	Note string `json:"note"`
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
	Year string `json:"year"`
}

func (item *Holiday) IsValidated() bool {
	if item.Name == "" {
		return false
	}
	if item.PartnerUid == "" {
		return false
	}
	if item.CourseUid == "" {
		return false
	}
	if item.Year == "" {
		return false
	}
	if item.From == "" {
		return false
	}
	if item.Name == "" {
		return false
	}
	return true
}

func (item *Holiday) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Holiday) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Holiday) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Holiday) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Holiday{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Holiday) FindList() ([]HolidayResponse, int64, error) {
	db := datasources.GetDatabase().Model(Holiday{})
	list := []HolidayResponse{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""

	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.Year != "" {
		db = db.Where("year = ?", item.Year)
	}
	db.Count(&total)
	db = db.Find(&list)
	return list, total, db.Error
}

func (item *Holiday) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
