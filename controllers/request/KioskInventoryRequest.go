package request

import (
	"database/sql/driver"
	"encoding/json"
)

type KioskInventoryInputItemBody struct {
	Code       string  `json:"code" binding:"required"`
	ItemCode   string  `json:"item_code" binding:"required"`
	ItemName   string  `json:"item_name" binding:"required"`
	Unit       string  `json:"unit" binding:"required"`
	GroupCode  string  `json:"group_code" binding:"required"`
	Quantity   int64   `json:"quantity" binding:"required"`
	UserUpdate string  `json:"user_update" binding:"required"`
	Note       string  `json:"note"`
	Price      float64 `json:"price" binding:"required"`
}

type ListKioskInventoryInputItemBody []KioskInventoryInputItemBody

func (item *ListKioskInventoryInputItemBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListKioskInventoryInputItemBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type KioskInventoryOutputItemBody struct {
	Code       string  `json:"code" binding:"required"`
	ItemCode   string  `json:"item_code" binding:"required"`
	ItemName   string  `json:"item_name" binding:"required"`
	Unit       string  `json:"unit" binding:"required"`
	GroupCode  string  `json:"group_code" binding:"required"`
	Quantity   int64   `json:"quantity" binding:"required"`
	UserUpdate string  `json:"user_update" binding:"required"`
	Note       string  `json:"note"`
	Price      float64 `json:"price" binding:"required"`
}

type ListKioskInventoryOutputItemBody []KioskInventoryInputItemBody

func (item *ListKioskInventoryOutputItemBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListKioskInventoryOutputItemBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type CreateInputBillBody struct {
	PartnerUid        string                          `json:"partner_uid" binding:"required"`
	CourseUid         string                          `json:"course_uid" binding:"required"`
	BillCode          string                          `json:"bill_code" binding:"required"`
	ServiceId         int64                           `json:"service_id" binding:"required"`
	ServiceName       string                          `json:"service_name" binding:"required"`
	ServiceExportId   int64                           `json:"service_export_id" binding:"required"`
	ServiceExportName string                          `json:"service_export_name" binding:"required"`
	ListItem          ListKioskInventoryInputItemBody `json:"list_item" binding:"required"`
	Note              string                          `json:"note"`
}

type CreateOutputBillBody struct {
	PartnerUid        string                           `json:"partner_uid" binding:"required"`
	CourseUid         string                           `json:"course_uid" binding:"required"`
	BillCode          string                           `json:"bill_code" binding:"required"`
	ServiceId         int64                            `json:"service_id" binding:"required"`
	ServiceName       string                           `json:"service_name" binding:"required"`
	ServiceImportId   int64                            `json:"service_import_id" binding:"required"`
	ServiceImportName string                           `json:"service_import_name" binding:"required"`
	ListItem          ListKioskInventoryOutputItemBody `json:"list_item" binding:"required"`
}

type GetInOutItems struct {
	PageRequest
	ServiceId  int64  `form:"service_id" binding:"required"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	ItemCode   string `form:"item_code"`
}

type GetBill struct {
	PageRequest
	BillStatus string `form:"bill_status"`
	ServiceId  int64  `form:"service_id" binding:"required"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
}

type KioskInventoryInsertBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Code       string `json:"code" binding:"required"` // Mã đơn nhập
	ServiceId  int64  `json:"service_id" binding:"required"`
}
