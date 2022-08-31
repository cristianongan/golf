package kiosk_inventory

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

/*
Lưu thông tin đơn nhập kho
*/
type InventoryBill struct {
	models.ModelId
	PartnerUid string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code       string `json:"code" gorm:"type:varchar(100);index"`        // mã nhập kho
	Source     string `json:"source"`                                     // nguồn từ đâu: từ kho tổng hay từ kiosk khác..?
	BillStatus string `json:"bill_status"`
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
