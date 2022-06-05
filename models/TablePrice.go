package models

import (
	"start/constants"
	"start/datasources"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Bảng phí
type TablePrice struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // Sân Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`              // Tên Bảng phí
	FromDate   int64  `json:"from_date" gorm:"index"`                     // Áp dụng phí này từ thời gian
}

func (item *TablePrice) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	if item.ModelId.Status == "" {
		item.ModelId.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *TablePrice) Update() error {
	mydb := datasources.GetDatabase()
	item.ModelId.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *TablePrice) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *TablePrice) Count() (int64, error) {
	db := datasources.GetDatabase().Model(TablePrice{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *TablePrice) FindList(page Page) ([]TablePrice, int64, error) {
	db := datasources.GetDatabase().Model(TablePrice{})
	list := []TablePrice{}
	total := int64(0)
	status := item.ModelId.Status
	item.ModelId.Status = ""
	// db = db.Where(item)
	if status != "" {
		db = db.Where("status in (?)", strings.Split(status, ","))
	}
	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *TablePrice) FindCurrentUse() (TablePrice, error) {
	db := datasources.GetDatabase().Model(TablePrice{})
	list := []TablePrice{}

	db = db.Where("partner_uid = ?", item.PartnerUid)
	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	err := db.Find(&list).Error

	if err != nil {
		return TablePrice{}, err
	}

	if len(list) == 0 {
		return TablePrice{}, errors.New("list empty")
	}

	maxFromDate := int64(0)
	indexCurrent := -1

	currentTime := time.Now().Unix()

	// Lấy theo điều kiện
	// TODO: điều kiện ap dụng bảng giá
	/*
		max from-date: ngày áp dụng
		max ngày update:
		ngày áp dụng T+1: fromDate > current + 1 ngày
	*/

	for i, v := range list {
		if v.Status == constants.STATUS_ENABLE {
			if v.FromDate > int64(maxFromDate) && currentTime > (v.FromDate+86400) {
				maxFromDate = v.FromDate
				indexCurrent = i
			}
		}
	}

	if indexCurrent == -1 {
		return TablePrice{}, errors.New("Not found table price valid")
	}

	return list[indexCurrent], nil
}

func (item *TablePrice) Delete() error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
