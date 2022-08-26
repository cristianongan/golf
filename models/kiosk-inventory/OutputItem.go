package kiosk_inventory

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"gorm.io/datatypes"
)

/*
 Để lưu thông tin xuất kho
*/
type InventoryOutputItem struct {
	models.ModelId
	Code       string         `json:"code"`        // Mã đơn xuất
	ItemCode   string         `json:"item_code"`   // mã của sản phẩm
	Quantity   int64          `json:"quantity"`    // số lượng
	OutputDate datatypes.Date `json:"output_date"` // ngày xuất kho
	Reason     string         `json:"reason"`      // lý do xuất kho
}

func (item *InventoryOutputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
