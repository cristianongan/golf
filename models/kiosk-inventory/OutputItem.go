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
	OutputDate  string   `json:"output_date"`                           // ngày xuất kho
	ServiceId   int64    `json:"service_id" gorm:"index"`               // mã service
	ServiceName string   `json:"service_name" gorm:"type:varchar(256)"` // tên service
}

type OutputStatisticItem struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
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
	db = db.Select("partner_uid, course_uid,item_code,SUM(quantity) as total").Group("partner_uid,course_uid,item_code")
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
