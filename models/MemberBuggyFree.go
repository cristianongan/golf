package models

import (
	"start/datasources"

	"github.com/pkg/errors"
)

// Loại khách hàng
type MemberBuggyFree struct {
	Id           int64  `gorm:"AUTO_INCREMENT:yes" sql:"bigint;not null;primary_key"  json:"id"`
	MemberCardId string `json:"member_card_id" gorm:"type:varchar(50)"` // VAT
}

func (item *MemberBuggyFree) Create() error {
	db := datasources.GetDatabaseAuth()
	return db.Create(item).Error
}

func (item *MemberBuggyFree) Update() error {
	db := datasources.GetDatabaseAuth()
	errUpdate := db.Save(item).Error
	if errUpdate != nil {
		return errUpdate
	}
	return nil
}

func (item *MemberBuggyFree) FindFirst() error {
	db := datasources.GetDatabaseAuth()
	return db.Where(item).First(item).Error
}

func (item *MemberBuggyFree) Count() (int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(MemberBuggyFree{})
	total := int64(0)
	db = db.Where(item)
	db = db.Count(&total)
	return total, db.Error
}

func (item *MemberBuggyFree) FindAll() ([]MemberBuggyFree, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(MemberBuggyFree{})
	list := []MemberBuggyFree{}
	db.Find(&list)
	return list, db.Error
}

func (item *MemberBuggyFree) FindList(page Page) ([]MemberBuggyFree, int64, error) {
	database := datasources.GetDatabaseAuth()
	db := database.Model(MemberBuggyFree{})
	list := []MemberBuggyFree{}
	total := int64(0)
	db = db.Where(item)
	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}
	return list, total, db.Error
}

func (item *MemberBuggyFree) Delete() error {
	db := datasources.GetDatabaseAuth()
	if item.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}
