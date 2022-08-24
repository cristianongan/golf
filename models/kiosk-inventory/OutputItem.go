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
	ItemCode   string         `json:"item_code"`
	Quantity   int64          `json:"quantity"`
	OutputDate datatypes.Date `json:"output_date"`
	Reason     string         `json:"reason"`
}

func (item *InventoryOutputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
