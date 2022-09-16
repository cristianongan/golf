package kiosk_inventory

import (
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

/*
 Để lưu thông tin xuất kho
*/
type InventoryOutputItem struct {
	models.ModelId
	PartnerUid  string   `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string   `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code        string   `json:"code" gorm:"type:varchar(100);index"`        // Mã đơn xuất
	ItemCode    string   `json:"item_code" gorm:"type:varchar(100);index"`   // mã của sản phẩm
	ItemInfo    ItemInfo `json:"item_info" gorm:"type:json"`
	Quantity    int64    `json:"quantity"`                              // số lượng
	Amount      int64    `json:"amount"`                                // Tổng tiền
	OutputDate  string   `json:"output_date"`                           // ngày xuất kho
	ServiceId   int64    `json:"service_id" gorm:"index"`               // mã service
	ServiceName string   `json:"service_name" gorm:"type:varchar(256)"` // tên service
}
type InventoryOutputItemWithBill struct {
	InventoryOutputItem
	ServiceImportId   int64  `json:"service_import_id"`
	ServiceImportName string `json:"service_import_name"`
	Bag               string `json:"bag"`
	CustomerName      string `json:"customer_name"`
	BillStatus        string `json:"bill_status"`
	UserUpdate        string `json:"user_update"`
}
type OutputStatisticItem struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	ServiceId  int64  `json:"service_id"`
	ItemCode   string `json:"item_code"`
	Total      int64  `json:"total"`
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

func (item *InventoryOutputItem) FindStatistic() ([]OutputStatisticItem, error) {
	db := datasources.GetDatabase().Model(InventoryOutputItem{})
	db = db.Select("partner_uid, course_uid,item_code,service_id,SUM(quantity) as total").Group("partner_uid,course_uid,item_code")
	if item.OutputDate != "" {
		db = db.Where("output_date = ?", item.OutputDate)
	}
	list := []OutputStatisticItem{}
	db.Find(&list)

	return list, db.Error
}

func (item *InventoryOutputItem) FindList(page models.Page) ([]InventoryOutputItem, int64, error) {
	db := datasources.GetDatabase().Model(InventoryOutputItem{})
	list := []InventoryOutputItem{}
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

func (item *InventoryOutputItem) FindListForStatistic(page models.Page, fromDate int64, toDate int64) ([]InventoryOutputItemWithBill, int64, error) {
	db := datasources.GetDatabase().Model(InventoryOutputItem{})
	db = db.Joins("JOIN output_inventory_bills on output_inventory_bills.code = inventory_output_items.code")
	db = db.Select("inventory_output_items.*,output_inventory_bills.service_import_id," +
		"output_inventory_bills.service_import_name,output_inventory_bills.bag," +
		"output_inventory_bills.customer_name,output_inventory_bills.bill_status,output_inventory_bills.user_update")
	list := []InventoryOutputItemWithBill{}
	total := int64(0)

	if item.PartnerUid != "" {
		db = db.Where("inventory_output_items.partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("inventory_output_items.course_uid = ?", item.CourseUid)
	}

	if item.Code != "" {
		db = db.Where("inventory_output_items.code = ?", item.Code)
	}

	if item.ServiceId > 0 {
		db = db.Where("inventory_output_items.service_id = ?", item.ServiceId)
	}

	if item.ItemCode != "" {
		db = db.Where("inventory_output_items.item_code = ?", item.ItemCode)
	}

	if fromDate > 0 {
		db = db.Where("output_inventory_bills.output_date >= ?", fromDate)
	}

	if toDate > 0 {
		db = db.Where("output_inventory_bills.output_date <= ?", toDate)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
