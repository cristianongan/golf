package models

import (
	"start/constants"
	"start/utils"

	"gorm.io/gorm"
)

type CaddieDeleted struct {
	Caddie
}

func (item *CaddieDeleted) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}
