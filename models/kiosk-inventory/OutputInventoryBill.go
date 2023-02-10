package kiosk_inventory

import (
	"start/constants"
	"start/models"
	"start/utils"

	"gorm.io/gorm"
)

/*
Lưu thông tin đơn nhập kho
*/
type OutputInventoryBill struct {
	models.ModelId
	PartnerUid        string                `json:"partner_uid" gorm:"type:varchar(100);index"`   // Hang Golf
	CourseUid         string                `json:"course_uid" gorm:"type:varchar(256);index"`    // San Golf
	Code              string                `json:"code" gorm:"type:varchar(100);index"`          // mã xuất kho
	OutputDate        int64                 `json:"output_date"`                                  // ngày xuất kho
	UserUpdate        string                `json:"user_update" gorm:"type:varchar(256)"`         // Người update cuối cùng
	ServiceId         int64                 `json:"service_id" gorm:"index"`                      // mã service
	ServiceName       string                `json:"service_name" gorm:"type:varchar(256)"`        // tên service
	Note              string                `json:"note" gorm:"type:varchar(256)"`                // ghi chú
	ServiceImportId   int64                 `json:"service_import_id"`                            // id service sẽ import
	ServiceImportName string                `json:"service_import_name" gorm:"type:varchar(256)"` // tên service import
	Bag               string                `json:"bag" gorm:"type:varchar(100);index"`           // Golf Bag
	CustomerName      string                `json:"customer_name" gorm:"type:varchar(256)"`       // Tên khách hàng chơi golf
	BillStatus        string                `json:"bill_status" gorm:"type:varchar(100)"`         // Trạng thái đơn hàng (SELL, TRANSFER)
	BillType          string                `json:"bill_type" gorm:"type:varchar(50)"`            // Trạng thái đơn hàng (SELL, TRANSFER)
	Quantity          int64                 `json:"quantity"`                                     // Tổng số lượng sell or transfer
	ListItem          []InventoryOutputItem `json:"list_item,omitempty" gorm:"foreignKey:Code;references:Code"`
}

func (item *OutputInventoryBill) IsDuplicated(db *gorm.DB) bool {
	bill := OutputInventoryBill{
		PartnerUid: item.PartnerUid,
		CourseUid:  item.CourseUid,
		ServiceId:  item.ServiceId,
		Code:       item.Code,
	}

	errFind := bill.FindFirst(db)
	if errFind == nil || bill.Id > 0 {
		return true
	}
	return false
}

func (item *OutputInventoryBill) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *OutputInventoryBill) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	return db.Save(item).Error
}
func (item *OutputInventoryBill) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}
func (item *OutputInventoryBill) FindList(database *gorm.DB, page models.Page, status string) ([]OutputInventoryBill, int64, error) {
	db := database.Model(OutputInventoryBill{})
	list := []OutputInventoryBill{}
	total := int64(0)

	if item.Code != "" {
		db = db.Where("code = ?", item.Code)
	}

	if item.PartnerUid != "" {
		db = db.Where("partner_uid = ?", item.PartnerUid)
	}

	if item.CourseUid != "" {
		db = db.Where("course_uid = ?", item.CourseUid)
	}

	if status != "" {
		db = db.Where("bill_status = ?", status)
	}

	if item.ServiceId > 0 {
		db = db.Where("service_id = ?", item.ServiceId)
	}

	db.Count(&total)
	if item.Code != "" {
		db = db.Preload("ListItem")
	}

	if total > 0 && int64(page.Offset()) < total {
		db = page.Setup(db).Find(&list)
	}

	return list, total, db.Error
}
