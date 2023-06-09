package kiosk_inventory

import (
	"database/sql/driver"
	"encoding/json"
	"start/constants"
	"start/models"
	"start/utils"

	"gorm.io/gorm"
)

/*
Để lưu thông tin nhập kho
*/
type InventoryInputItem struct {
	models.ModelId
	PartnerUid  string   `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string   `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code        string   `json:"code" gorm:"type:varchar(100);index"`        // mã nhập kho
	ItemCode    string   `json:"item_code" gorm:"type:varchar(100);index"`   // mã sản phẩm
	ItemInfo    ItemInfo `json:"item_info" gorm:"type:json"`
	ServiceId   int64    `json:"service_id" gorm:"index"`               // mã service
	ServiceName string   `json:"service_name" gorm:"type:varchar(256)"` // tên service
	Quantity    int64    `json:"quantity"`                              // số lượng
	Amount      int64    `json:"amount"`                                // Tổng tiền
	InputDate   string   `json:"input_date"`                            // ngày nhập kho
}
type InventoryInputItemWithBill struct {
	InventoryInputItem
	ServiceExportId   int64  `json:"service_export_id"`
	ServiceExportName string `json:"service_export_name"`
	BillStatus        string `json:"bill_status"`
	UserUpdate        string `json:"user_update"`
	Note              string `json:"note"`
}
type ItemInfo struct {
	Price     float64 `json:"price"`                               // Giá sản phẩm
	ItemName  string  `json:"item_name" gorm:"type:varchar(256)"`  // Tên sản phẩm
	GroupName string  `json:"group_name" gorm:"type:varchar(100)"` // Group Name
	GroupType string  `json:"group_type" gorm:"type:varchar(100)"` // Group Type
	GroupCode string  `json:"group_code" gorm:"type:varchar(100)"` // Group Type
	Unit      string  `json:"unit" gorm:"type:varchar(100)"`       // Đơn vị
}

func (item *ItemInfo) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ItemInfo) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *InventoryInputItem) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *InventoryInputItem) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()
	return db.Save(item).Error
}

func (item *InventoryInputItem) FindAllList(database *gorm.DB) ([]InventoryInputItem, int64, error) {
	db := database.Model(InventoryInputItem{})
	list := []InventoryInputItem{}
	total := int64(0)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	db.Count(&total)
	db.Find(&list)

	return list, total, db.Error
}

func (item *InventoryInputItem) FindList(database *gorm.DB, page models.Page, itemType string) ([]InventoryInputItem, int64, error) {
	db := database.Model(InventoryInputItem{})
	list := []InventoryInputItem{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	if item.ServiceId > 0 {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	if item.ItemCode != "" {
		db = db.Where("item_code = ?", item.ItemCode)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *InventoryInputItem) FindStatistic(database *gorm.DB) ([]OutputStatisticItem, error) {
	db := database.Model(InventoryInputItem{})
	db = db.Joins("JOIN input_inventory_bills ON input_inventory_bills.code = inventory_input_items.code")
	db = db.Where("input_inventory_bills.bill_status = ?", "APPROVED")
	db = db.Group("inventory_input_items.item_code")
	db = db.Select("inventory_input_items.partner_uid,inventory_input_items.course_uid,inventory_input_items.service_id,inventory_input_items.item_code,SUM(inventory_input_items.quantity) as total")
	if item.InputDate != "" {
		db = db.Where("inventory_input_items.input_date = ?", item.InputDate)
	}
	list := []OutputStatisticItem{}
	db.Find(&list)

	return list, db.Error
}

func (item *InventoryInputItem) FindListForStatistic(database *gorm.DB, page models.Page, fromDate int64, toDate int64) ([]InventoryInputItemWithBill, int64, error) {
	db := database.Model(InventoryInputItem{})
	db = db.Joins("JOIN input_inventory_bills on input_inventory_bills.code = inventory_input_items.code")
	db = db.Select("inventory_input_items.*,input_inventory_bills.service_export_id," +
		"input_inventory_bills.service_export_name,input_inventory_bills.bill_status," +
		"input_inventory_bills.user_update,input_inventory_bills.note")
	list := []InventoryInputItemWithBill{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("inventory_input_items.partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("inventory_input_items.course_uid = ?", item.CourseUid)
	}

	if item.Code != "" {
		db = db.Where("inventory_input_items.code = ?", item.Code)
	}

	if item.ServiceId > 0 {
		db = db.Where("inventory_input_items.service_id = ?", item.ServiceId)
	}

	if item.ItemCode != "" {
		db = db.Where("inventory_input_items.item_code = ?", item.ItemCode)
	}

	if fromDate > 0 {
		db = db.Where("input_inventory_bills.input_date >= ?", fromDate)
	}

	if toDate > 0 {
		db = db.Where("input_inventory_bills.input_date <= ?", toDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
