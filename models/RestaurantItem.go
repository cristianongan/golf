package models

import (
	"start/constants"
	"start/utils"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

/*
Giỏ Hàng
*/
type RestaurantItem struct {
	ModelId
	PartnerUid       string `json:"partner_uid" gorm:"type:varchar(100);index"` // Hãng golf
	CourseUid        string `json:"course_uid" gorm:"type:varchar(150);index"`  // Sân golf
	ServiceId        int64  `json:"service_id" gorm:"index"`                    // Mã của service
	OrderDate        string `json:"order_date" gorm:"type:varchar(30);index"`   // Ex: 06/11/2022
	Type             string `json:"type" gorm:"type:varchar(100)"`              // Loại sản phẩm: FOOD, DRINK
	BillId           int64  `json:"bill_id" gorm:"index"`                       // id hóa đơn
	ItemId           int64  `json:"item_id" gorm:"index"`                       // id sản phẩm
	ItemCode         string `json:"item_code" gorm:"type:varchar(100)"`         // Mã sản phẩm
	ItemName         string `json:"item_name" gorm:"type:varchar(100)"`         // Tên sản phẩm
	ItemComboCode    string `json:"item_combo_code" gorm:"type:varchar(100)"`   // Code combo
	ItemComboName    string `json:"item_combo_name" gorm:"type:varchar(100)"`   // Tên combo
	ItemUnit         string `json:"item_unit" gorm:"type:varchar(100)"`         // Đơn vị
	ItemStatus       string `json:"item_status" gorm:"type:varchar(100)"`       // Trạng thái sản phẩm
	ItemNote         string `json:"item_note" gorm:"type:varchar(200)"`         // Yêu cầu của khách hàng
	Quantity         int    `json:"quantity"`                                   // Số lượng order
	QuantityProgress int    `json:"quantity_progress"`                          // Số lương đang tiến hành
	TotalProcess     int    `json:"total_process"`                              // Tổng số lượng đang làm
	MoveKitchenTimes int    `json:"move_kitchen_times"`                         // Số lần move kitchen của bill
}

func (item *RestaurantItem) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()

	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *RestaurantItem) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}

func (item *RestaurantItem) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	return db.Save(item).Error
}

func (item *RestaurantItem) UpdateBatchBillId(db *gorm.DB) error {
	db = db.Model(RestaurantItem{})

	if item.ItemId != 0 {
		db = db.Where("item_id = ?", item.ItemId)
	}

	if item.BillId != 0 {
		db = db.Update("bill_id", item.BillId)
	}

	return db.Error
}

func (item *RestaurantItem) Delete(db *gorm.DB) error {
	if item.ModelId.Id <= 0 {
		return errors.New("Primary key is undefined!")
	}
	return db.Delete(item).Error
}

func (item *RestaurantItem) FindList(database *gorm.DB, page Page) ([]RestaurantItem, int64, error) {
	var list []RestaurantItem
	total := int64(0)

	db := database.Model(RestaurantItem{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.ServiceId != 0 {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	if item.Id != 0 {
		db = db.Where("id = ?", item.Id)
	}

	db = db.Where("order_date = ?", item.OrderDate)

	db.Count(&total)

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}

func (item *RestaurantItem) FindAll(database *gorm.DB) ([]RestaurantItem, error) {
	var list []RestaurantItem

	db := database.Model(RestaurantItem{})

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if item.BillId != 0 {
		db = db.Where("bill_id = ?", item.BillId)
	}

	if item.Id != 0 {
		db = db.Where("id = ?", item.Id)
	}

	if item.ServiceId != 0 {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	if item.ItemCode != "" {
		db = db.Where("item_code = ?", item.ItemCode)
	}

	if item.ItemId != 0 {
		db = db.Where("item_id = ?", item.ItemId)
	}

	if item.ItemStatus != "" {
		db = db.Where("item_status = ?", item.ItemStatus)
	}

	db.Find(&list)

	return list, db.Error
}

func (item *RestaurantItem) FindAllGroupBy(database *gorm.DB) ([]map[string]interface{}, error) {
	db := database.Table("restaurant_items")
	var list []map[string]interface{}

	db = db.Select("restaurant_items.*", "service_carts.*")

	if item.CourseUid != "" {
		db = db.Where("restaurant_items.course_uid = ?", item.CourseUid)
	}
	if item.PartnerUid != "" {
		db = db.Where("restaurant_items.partner_uid = ?", item.PartnerUid)
	}
	if item.ServiceId != 0 {
		db = db.Where("restaurant_items.service_id = ?", item.ServiceId)
	}
	if item.Type != "" {
		db = db.Where("restaurant_items.type = ?", item.Type)
	}
	if item.ItemStatus != "" {
		db = db.Where("restaurant_items.item_status = ?", item.ItemStatus)
	}
	if item.OrderDate != "" {
		db = db.Where("restaurant_items.order_date = ?", item.OrderDate)
	}

	db = db.Joins("INNER JOIN service_carts on service_carts.id = restaurant_items.bill_id")

	// db = db.Where("item_status = ?", constants.RES_STATUS_PROCESS)
	db.Order("service_carts.time_process desc")

	db.Find(&list)

	return list, db.Error
}
