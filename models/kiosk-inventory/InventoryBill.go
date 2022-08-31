package kiosk_inventory

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"

	"gorm.io/datatypes"
)

/*
Lưu thông tin đơn nhập kho
*/
type InventoryBill struct {
	models.ModelId
	PartnerUid string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string         `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code       string         `json:"code" gorm:"type:varchar(100);index"`        // mã nhập kho
	Source     string         `json:"source"`                                     // nguồn từ đâu: từ kho tổng hay từ kiosk khác..?
	BillStatus string         `json:"bill_status"`
	Note       string         `json:"note"`                                 // ghi chú
	InputDate  datatypes.Date `json:"input_date"`                           // ngày nhập kho
	UserUpdate string         `json:"user_update" gorm:"type:varchar(256)"` // Người update cuối cùngUserUpdate
	Type       string         `json:"type"`                                 // Loại bill (IMPORT, EXPORT)
}

func (item *InventoryBill) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *InventoryBill) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}
func (item *InventoryBill) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}
func (item *InventoryBill) FindList(page models.Page, status string) ([]InventoryInputItemResponse, int64, error) {
	db := datasources.GetDatabase().Model(InventoryBill{})
	list := []InventoryInputItemResponse{}
	total := int64(0)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	if status != "" {
		db = db.Where("bill_status = ?", status)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
