package kiosk_inventory

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

type InventoryItem struct {
	models.ModelId
	Code     string `json:"code"`
	Name     string `json:"name"`
	Quantity int64  `json:"quantity"`
}

func (item *InventoryItem) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *InventoryItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *InventoryItem) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}
