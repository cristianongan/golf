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
	ItemInfo   ItemInfo       `json:"item_info" gorm:"type:json"`
	Quantity   int64          `json:"quantity"`                                  // số lượng
	OutputDate datatypes.Date `json:"output_date"`                               // ngày xuất kho
	KioskCode  string         `json:"kiosk_code" gorm:"type:varchar(100);index"` // mã kiosk
	KioskName  string         `json:"kiosk_name" gorm:"type:varchar(256)"`       // tên kiosk
	Reason     string         `json:"reason" gorm:"type:varchar(256)"`           // lý do xuất kho
}

func (item *InventoryOutputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *InventoryOutputItem) FindAllList() ([]InventoryOutputItem, int64, error) {
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

func (item *InventoryOutputItem) FindList(page models.Page) ([]InventoryOutputItem, int64, error) {
	db := datasources.GetDatabase().Model(InventoryOutputItem{})
	list := []InventoryOutputItem{}
	total := int64(0)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
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
