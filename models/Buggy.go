package models

import (
	"github.com/pkg/errors"
	"start/constants"
	"start/datasources"
	"time"
)

type Buggy struct {
	ModelId
	CourseId string `json:"course_id" gorm:"type:varchar(100);index"`
	Number   int    `json:"number" gorm:"type:int"`
	Origin   string `json:"origin" gorm:"type:varchar(200)"`
	Note     string `json:"note" gorm:"type:varchar(200)"`
}

type BuggyResponse struct {
	ModelId
	CourseId string `json:"course_id"`
	Number   int    `json:"number"`
	Origin   string `json:"origin"`
	Note     string `json:"note"`
}

func (item *Buggy) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Buggy) Delete() error {
	if item.ModelId.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}

func (item *Buggy) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Buggy) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Buggy) Count() (int64, error) {
	total := int64(0)

	db := datasources.GetDatabase().Model(Buggy{})
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Buggy) FindList(page Page) ([]Buggy, int64, error) {
	var list []Buggy
	total := int64(0)

	db := datasources.GetDatabase().Model(Buggy{})
	db = db.Where(item)
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}
