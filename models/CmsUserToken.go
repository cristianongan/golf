package models

import (
	"errors"
	"start/constants"
	"start/datasources"
	"strings"
	"time"
)

type CmsUserToken struct {
	ModelId
	UserUid    string `json:"user_uid" gorm:"index;type:varchar(100)"`
	PartnerUid string `json:"partner_uid" gorm:"index;type:varchar(100)"`
	UserName   string `json:"user_name" gorm:"type:varchar(20)"`
	Token      string `json:"token"`
}

// ======= CRUD ===========
func (item *CmsUserToken) Create() error {
	now := time.Now()
	item.CreatedAt = now.Unix()
	item.UpdatedAt = now.Unix()
	item.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CmsUserToken) Update() error {
	mydb := datasources.GetDatabase()
	item.UpdatedAt = time.Now().Unix()
	errUpdate := mydb.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *CmsUserToken) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CmsUserToken) Count() (int64, error) {
	db := datasources.GetDatabase().Model(CmsUserToken{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *CmsUserToken) FindList(page Page) ([]CmsUserToken, int64, error) {
	db := datasources.GetDatabase().Model(CmsUserToken{})
	list := []CmsUserToken{}
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

func (item *CmsUserToken) Delete() error {
	if item.Id == 0 {
		return errors.New("Primary key is undefined!")
	}
	return datasources.GetDatabase().Delete(item).Error
}
