package kiosk_inventory

import (
	"gorm.io/datatypes"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

type InventoryInputItem struct {
	models.ModelId
	Code      string         `json:"code"`
	Quantity  int64          `json:"quantity"`
	InputDate datatypes.Date `json:"input_date"`
}

func (item *InventoryInputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
