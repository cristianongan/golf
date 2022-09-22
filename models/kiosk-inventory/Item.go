package kiosk_inventory

import (
	"log"
	"start/constants"
	"start/datasources"
	"start/models"
	"time"
)

/*
 Để lưu thông tin trạng thái item trong kho(trạng thái tồn)
*/
type InventoryItem struct {
	models.ModelId
	PartnerUid  string   `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string   `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	ServiceId   int64    `json:"service_id" gorm:"index"`                    // mã service
	ServiceName string   `json:"service_name" gorm:"type:varchar(256)"`      // tên service
	Code        string   `json:"code" gorm:"type:varchar(100)"`              // mã sp
	ItemInfo    ItemInfo `json:"item_info" gorm:"type:json"`                 // Thông tin sản phầm
	Quantity    int64    `json:"quantity"`                                   // số lượng
	StockStatus string   `json:"stock_status" gorm:"type:varchar(100)"`      // trạng thái: còn hàng hay hết hàng
}

type InventoryItemRequest struct {
	ServiceId   int64
	PartnerUid  string
	CourseUid   string
	ItemCode    string
	FromDate    string
	ToDate      string
	Type        string
	ProductName string
}

func (item *InventoryItem) FindFirst() error {
	db := datasources.GetDatabase()
	return db.Where(item).First(item).Error
}

func (item *InventoryItem) Create() error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	db := datasources.GetDatabase()
	return db.Create(item).Error
}

func (item *InventoryItem) Update() error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	db := datasources.GetDatabase()
	return db.Save(item).Error
}

/// ------- InventoryItem batch insert to db ------
func (item *InventoryItem) BatchInsert(list []InventoryItem) error {
	db := datasources.GetDatabase().Table("inventory_items")
	var err error
	err = db.Create(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch insert err: ", err.Error())
	}
	return err
}
func (item *InventoryItem) FindList(page models.Page, param InventoryItemRequest) ([]InventoryItem, int64, error) {
	db := datasources.GetDatabase().Model(InventoryItem{})
	list := []InventoryItem{}
	total := int64(0)

	if param.PartnerUid != "" {
		db = db.Where("partner_uid = ?", param.PartnerUid)
	}

	if param.CourseUid != "" {
		db = db.Where("course_uid = ?", param.CourseUid)
	}

	if param.ServiceId > 0 {
		db = db.Where("service_id = ?", param.ServiceId)
	}

	if param.ItemCode != "" {
		db = db.Where("code = ?", param.ItemCode)
	}

	if param.Type != "" {
		db = db.Where("item_info->'$.group_type' = ?", param.Type)
	}

	if param.ProductName != "" {
		db = db.Where("item_info->'$.item_name' = ?", param.ProductName)
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
