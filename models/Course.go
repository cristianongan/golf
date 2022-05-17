package models

import (
	"errors"
	"start/constants"
	"start/datasources"
	"strings"
	"time"
)

// Sân Golf
type Course struct {
	Model
	PartnerUid  string  `json:"partner_uid" gorm:"type:varchar(100);index"`
	Name        string  `json:"name" gorm:"type:varchar(256)"`
	Hole        int     `json:"hole"`
	Address     string  `json:"address" gorm:"type:varchar(500)"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Icon        string  `json:"icon" gorm:"type:varchar(256)"`
	RateGolfFee string  `json:"rate_golf_fee" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *Course) Create() error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()
	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Course) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Course) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Course) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Course{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Course) FindList(page Page) ([]Course, int64, error) {
	db := datasources.GetDatabase().Model(Course{})
	list := []Course{}
	total := int64(0)
	status := item.Status
	item.Status = ""
	db = db.Where(item)

	if status != "" {
		db = db.Where("status IN (?)", strings.Split(status, ","))
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *Course) Delete() error {
	if item.Uid == "" {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
