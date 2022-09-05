package kiosk_inventory

import (
	"database/sql/driver"
	"encoding/json"
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
	ItemInfo      ItemInfo       `json:"item_info" gorm:"type:json"`
	KioskCode     string         `json:"kiosk_code" gorm:"type:varchar(100);index"` // mã kiosk
	KioskName     string         `json:"kiosk_name" gorm:"type:varchar(256)"`       // tên kiosk
	Quantity      int64          `json:"quantity"`                                  // số lượng
	InputDate     datatypes.Date `json:"input_date"`                                // ngày nhập kho
	Source        string         `json:"source" gorm:"type:varchar(100)"`           // nguồn từ đâu: từ kho tổng hay từ kiosk khác..?
	ReviewUserUid string         `json:"review_user_uid" gorm:"type:varchar(256)"`  // Người duyệt khi nhập kho
	Note          string         `json:"note" gorm:"type:varchar(256)"`             // ghi chú
}

type ItemInfo struct {
	Price     float64 `json:"price"`                               // Giá sản phẩm
	ItemName  string  `json:"item_name" gorm:"type:varchar(256)"`  // mã kiosk
	GroupName string  `json:"group_name" gorm:"type:varchar(100)"` // mã kiosk
	GroupType string  `json:"group_type" gorm:"type:varchar(100)"` // mã kiosk
}

func (item *ItemInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ItemInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *InventoryInputItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}
func (item *InventoryInputItem) FindAllList() ([]InventoryInputItem, int64, error) {
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

func (item *InventoryInputItem) FindList(page models.Page) ([]InventoryInputItem, int64, error) {
	db := datasources.GetDatabase().Model(InventoryInputItem{})
	list := []InventoryInputItem{}
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
