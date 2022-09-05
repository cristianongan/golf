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
type OutputInventoryBill struct {
	models.ModelId
	PartnerUid string         `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid  string         `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code       string         `json:"code" gorm:"type:varchar(100);index"`        // mã xuất kho
	Source     string         `json:"source" gorm:"type:varchar(100)"`            // nguồn từ đâu: từ kho tổng hay từ kiosk khác..?
	BillStatus string         `json:"bill_status" gorm:"type:varchar(100)"`
	InputDate  datatypes.Date `json:"input_date"`                                // ngày nhập kho
	UserUpdate string         `json:"user_update" gorm:"type:varchar(256)"`      // Người update cuối cùng
	KioskCode  string         `json:"kiosk_code" gorm:"type:varchar(100);index"` // mã kiosk
	KioskName  string         `json:"kiosk_name" gorm:"type:varchar(256)"`       // tên kiosk
	Note       string         `json:"note" gorm:"type:varchar(256)"`             // ghi chú
}

func (item *OutputInventoryBill) IsDuplicated() bool {
	bill := OutputInventoryBill{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		KioskCode:  item.KioskCode,
		Code:       item.Code,
	}

	errFind := bill.FindFirst()
	if errFind == nil || bill.Id > 0 {
		return true
	}
	return false
}

func (item *OutputInventoryBill) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *OutputInventoryBill) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}
func (item *OutputInventoryBill) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}
func (item *OutputInventoryBill) FindList(page models.Page, status string) ([]OutputInventoryBill, int64, error) {
	db := datasources.GetDatabase().Model(OutputInventoryBill{})
	list := []OutputInventoryBill{}
	total := int64(0)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	if status != "" {
		db = db.Where("bill_status = ?", status)
	}

	if item.KioskCode != "" {
		db = db.Where("kiosk_code = ?", item.KioskCode)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
