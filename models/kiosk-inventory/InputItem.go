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
	Code          string         `json:"code"`
	ItemCode      string         `json:"item_code"`
	Quantity      int64          `json:"quantity"`
	InputDate     datatypes.Date `json:"input_date"`
	Source        string         `json:"source"`
	InputStatus   string         `json:"input_status"`
	ReviewUserUid string         `json:"review_user_uid"`
	Note          string         `json:"note"`
}

func (item *InventoryInputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
