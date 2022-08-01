package models

import (
	"start/constants"
	"start/datasources"
	"time"
)

type CaddieGroup struct {
	ModelId
	Name       string   `json:"name"`
	Code       string   `json:"code"`
	PartnerUid string   `json:"partner_uid"`
	CourseUid  string   `json:"course_uid"`
	Caddies    []Caddie `json:"caddies" gorm:"foreignKey:GroupId"`
}

func (item *CaddieGroup) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *CaddieGroup) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *CaddieGroup) Delete() error {
	return datasources.GetDatabase().Delete(item).Error
}

func (item *CaddieGroup) FindList(page Page) ([]CaddieGroup, int64, error) {
	var list []CaddieGroup
	total := int64(0)

	db := datasources.GetDatabase().Model(CaddieGroup{})

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Preload("Caddies").Find(&list)
	}
	return list, total, db.Error
}
