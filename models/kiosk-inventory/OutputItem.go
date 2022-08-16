package kiosk_inventory

import (
	"gorm.io/datatypes"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

type InventoryOutputItem struct {
	models.ModelId
	Code       string         `json:"code"`
	Quantity   int64          `json:"quantity"`
	OutputDate datatypes.Date `json:"output_date"`
}

func (item *InventoryOutputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
