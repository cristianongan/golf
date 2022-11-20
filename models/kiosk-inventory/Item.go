package kiosk_inventory

import (
	"log"
	"start/constants"
	"start/models"
	"time"

	"gorm.io/gorm"
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
	InStock     string
}

func (item *InventoryItem) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *InventoryItem) Create(db *gorm.DB) error {
	now := time.Now()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *InventoryItem) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = time.Now().Unix()

	return db.Save(item).Error
}

// / ------- InventoryItem batch insert to db ------
func (item *InventoryItem) BatchInsert(database *gorm.DB, list []InventoryItem) error {
	db := database.Table("inventory_items")
	var err error
	err = db.Create(&list).Error

	if err != nil {
		log.Println("BookingServiceItem batch insert err: ", err.Error())
	}
	return err
}
func (item *InventoryItem) FindList(database *gorm.DB, page models.Page, param InventoryItemRequest) ([]InventoryItem, int64, error) {
	db := database.Model(InventoryItem{})
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

	if param.ItemCode != "" || param.ProductName != "" {
		db = db.Where("code COLLATE utf8mb4_general_ci LIKE ? OR item_info->'$.item_name' COLLATE utf8mb4_general_ci LIKE ?", "%"+param.ItemCode+"%", "%"+param.ProductName+"%")
	}

	if param.Type != "" {
		if param.Type == constants.GROUP_FB_FOOD {
			db = db.Where("item_info->'$.group_type' = ?", "G1")
		} else if param.Type == constants.GROUP_FB_DRINK {
			db = db.Where("item_info->'$.group_type' = ?", "G2")
		}
	}

	if param.InStock != "" {
		if param.InStock == "1" {
			db.Where("quantity > 0")
		} else {
			db.Where("quantity = ?", 0)
		}
	}

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
