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
	PartnerUid  string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid   string `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	KioskCode   string `json:"kiosk_code"`                                 // mã kiosk
	InputCode   string `json:"input_code"`                                 // mã nhập kho
	Code        string `json:"code"`                                       // mã item
	Quantity    int64  `json:"quantity"`                                   // số lượng
	StockStatus string `json:"stock_status"`                               // trạng thái: còn hàng hay hết hàng
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
