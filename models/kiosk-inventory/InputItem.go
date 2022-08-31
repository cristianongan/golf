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
	PartnerUid    string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid     string         `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code          string         `json:"code" gorm:"type:varchar(100);index"`        // mã nhập kho
	ItemCode      string         `json:"item_code" gorm:"type:varchar(100);index"`   // mã sản phẩm
	KioskCode     string         `json:"kiosk_code"`                                 // mã kiosk
	Quantity      int64          `json:"quantity"`                                   // số lượng
	InputDate     datatypes.Date `json:"input_date"`                                 // ngày nhập kho
	Source        string         `json:"source"`                                     // nguồn từ đâu: từ kho tổng hay từ kiosk khác..?
	ReviewUserUid string         `json:"review_user_uid"`                            // Người duyệt khi nhập kho
	Note          string         `json:"note"`                                       // ghi chú
}

func (item *InventoryInputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
func (item *InventoryInputItem) FindList() ([]InventoryInputItem, int64, error) {
	db := datasources.GetDatabase().Model(InventoryInputItem{})
	list := []InventoryInputItem{}
	total := int64(0)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	db.Count(&total)
	db.Find(&list)

	return list, total, db.Error
}
