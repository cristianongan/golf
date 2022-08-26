package kiosk_inventory

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"gorm.io/datatypes"
)

/*
 Để lưu thông tin nhập kho
*/
type InventoryInputItem struct {
	models.ModelId
	Code          string         `json:"code"`            // mã nhập kho
	ItemCode      string         `json:"item_code"`       // mã sản phẩm
	Quantity      int64          `json:"quantity"`        // số lượng
	InputDate     datatypes.Date `json:"input_date"`      // ngày nhập kho
	Source        string         `json:"source"`          // nguồn từ đâu: từ kho tổng hay từ kiosk khác..?
	InputStatus   string         `json:"input_status"`    // hoàn , huỷ,..?
	ReviewUserUid string         `json:"review_user_uid"` // Người duyệt khi nhập kho
	Note          string         `json:"note"`            // ghi chú
}

func (item *InventoryInputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
