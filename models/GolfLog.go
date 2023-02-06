package models

import (
	"start/constants"
	"start/datasources"
	"start/utils"
	"strings"

	"github.com/pkg/errors"
)

// Logs
type GolfLog struct {
	ModelId
	Category string `json:"category" gorm:"type:varchar(200)"`
	Message  string `json:"message" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *GolfLog) Create() error {
	now := utils.GetTimeNow()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *GolfLog) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = utils.GetTimeNow().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *GolfLog) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *GolfLog) Count() (int64, error) {
	db := datasources.GetDatabase().Model(GolfLog{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *GolfLog) FindList(page Page) ([]GolfLog, int64, error) {
	db := datasources.GetDatabase().Model(GolfLog{})
	list := []GolfLog{}
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

func (item *GolfLog) Delete() error {
	if item.Id < 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
