package models

import (
	"errors"
	"start/constants"
	"start/datasources"
	"strings"
	"time"
)

// Company
type Company struct {
	ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Name       string `json:"name" gorm:"type:varchar(256)"`
}

// ======= CRUD ===========
func (item *Company) Create() error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()

	if item.Status == "" {
		item.Status = constants.STATUS_ENABLE
	}

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *Company) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *Company) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *Company) Count() (int64, error) {
	db := datasources.GetDatabase().Model(Company{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *Company) FindList(page Page) ([]Company, int64, error) {
	db := datasources.GetDatabase().Model(Company{})
	list := []Company{}
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

func (item *Company) Delete() error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
