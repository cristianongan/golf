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
	PartnerUid string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string         `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code       string         `json:"code" gorm:"type:varchar(100);index"`        // Mã đơn xuất
	ItemCode   string         `json:"item_code" gorm:"type:varchar(100);index"`   // mã của sản phẩm
	Quantity   int64          `json:"quantity"`                                   // số lượng
	OutputDate datatypes.Date `json:"output_date"`                                // ngày xuất kho
	Reason     string         `json:"reason"`                                     // lý do xuất kho
}

func (item *InventoryOutputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
func (item *InventoryOutputItem) FindList() ([]InventoryOutputItem, int64, error) {
	db := datasources.GetDatabase().Model(InventoryOutputItem{})
	list := []InventoryOutputItem{}
	total := int64(0)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	db.Count(&total)
	db.Find(&list)

	return list, total, db.Error
}
