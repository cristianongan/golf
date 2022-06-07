package models

import (
	"start/constants"
	"start/datasources"
	"time"

	"github.com/pkg/errors"
)

type CaddieAbsent struct {
	ModelId
	CourseId string `json:"course_id" gorm:"type:varchar(100);index"`
	CaddieId string `json:"caddie_id" gorm:"type:varchar(40)"`
	From     int64  `json:"from"`
	To       int64  `json:"to"`
	Type     string `json:"type" gorm:"type:varchar(20)"`
	Note     string `json:"note" gorm:"type:varchar(200)"`
}

type CaddieAbsentResponse struct {
	ModelId
	CourseId string `json:"course_id"`
	CaddieId string `json:"caddie_id"`
	From     int64  `json:"from"`
	To       int64  `json:"to"`
	Type     string `json:"type"`
	Note     string `json:"note"`
}

func (item *CaddieAbsent) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieAbsent) Delete() error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *CaddieAbsent) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CaddieAbsent) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieAbsent) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieAbsent{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CaddieAbsent) FindList(page Page, from int64, to int64) ([]CaddieAbsent, int64, error) {
	var list []CaddieAbsent
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieAbsent{})
	db = db.Where(item)

	if from > 0 {
		db = db.Where("created_at >= ?", from)
	}

	if to > 0 {
		db = db.Where("created_at < ?", to)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
