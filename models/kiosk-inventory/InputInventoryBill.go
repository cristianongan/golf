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
type InputInventoryBill struct {
	models.ModelId
	PartnerUid        string               `json:"partner_uid" gorm:"type:varchar(100);index"` // Hang Golf
	CourseUid         string               `json:"course_uid" gorm:"type:varchar(256);index"`  // San Golf
	Code              string               `json:"code" gorm:"type:varchar(100);index"`        // mã nhập kho
	BillStatus        string               `json:"bill_status" gorm:"type:varchar(100)"`
	Note              string               `json:"note" gorm:"type:varchar(256)"`                // ghi chú
	ServiceId         int64                `json:"service_id" gorm:"index"`                      // mã service
	ServiceName       string               `json:"service_name" gorm:"type:varchar(256)"`        // tên service
	ServiceExportId   int64                `json:"service_export_id"`                            // id service export
	ServiceExportName string               `json:"service_export_name" gorm:"type:varchar(256)"` // tên service export
	Quantity          int64                `json:"quantity"`                                     // Tổng số lượng sell or transfer
	UserUpdate        string               `json:"user_update" gorm:"type:varchar(256)"`         // Người update cuối cùng UserUpdate
	UserExport        string               `json:"user_export" gorm:"type:varchar(256)"`         // Người export đơn
	InputDate         int64                `json:"input_date"`                                   // ngày chấp nhận nhập kho
	OutputDate        int64                `json:"output_date"`                                  // ngày của đơn xuất
	ListItem          []InventoryInputItem `json:"list_item,omitempty" gorm:"foreignKey:Code;references:Code"`
}

func (item *InputInventoryBill) IsDuplicated(db *gorm.DB) bool {
	bill := InputInventoryBill{
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

func (item *InputInventoryBill) Create(db *gorm.DB) error {
	now := utils.GetTimeNow()
	item.ModelId.CreatedAt = now.Unix()
	item.ModelId.UpdatedAt = now.Unix()
	item.ModelId.Status = constants.STATUS_ENABLE

	return db.Create(item).Error
}

func (item *InputInventoryBill) Update(db *gorm.DB) error {
	item.ModelId.UpdatedAt = utils.GetTimeNow().Unix()

	return db.Save(item).Error
}
func (item *InputInventoryBill) FindFirst(db *gorm.DB) error {
	return db.Where(item).First(item).Error
}
func (item *InputInventoryBill) FindList(database *gorm.DB, page models.Page, status string) ([]InputInventoryBill, int64, error) {
	db := database.Model(InputInventoryBill{})
	list := []InputInventoryBill{}
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
