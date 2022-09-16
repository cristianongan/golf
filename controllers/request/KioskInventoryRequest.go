package request

import (
	"database/sql/driver"
	"encoding/json"
)

type KioskInventoryItemBody struct {
	ItemCode   string  `json:"item_code" binding:"required"`
	ItemName   string  `json:"item_name" binding:"required"`
	Unit       string  `json:"unit" binding:"required"`
	GroupCode  string  `json:"group_code" binding:"required"`
	Quantity   int64   `json:"quantity" binding:"required"`
	UserUpdate string  `json:"user_update" binding:"required"`
	Price      float64 `json:"price" binding:"required"`
}

type ListKioskInventoryInputItemBody []KioskInventoryItemBody

func (item *ListKioskInventoryInputItemBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListKioskInventoryInputItemBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type CreateBillBody struct {
	PartnerUid  string                          `json:"partner_uid" binding:"required"`
	CourseUid   string                          `json:"course_uid" binding:"required"`
	ServiceId   int64                           `json:"service_id" binding:"required"`
	ServiceName string                          `json:"service_name" binding:"required"`
	SourceId    int64                           `json:"source_id"`
	SourceName  string                          `json:"source_name"`
	ListItem    ListKioskInventoryInputItemBody `json:"list_item" binding:"required"`
	Note        string                          `json:"note"`
	UserExport  string                          `json:"user_export"`
	OutputDate  int64                           `json:"output_date"`
}

type CreateOutputBillBody struct {
	CreateBillBody
	Bag          string `json:"bag"`
	CustomerName string `json:"customer_name"`
}

type GetInOutItems struct {
	PageRequest
	ServiceId  int64  `form:"service_id" binding:"required"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	ItemCode   string `form:"item_code"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
}

type GetItems struct {
	PageRequest
	ServiceId  int64  `form:"service_id" binding:"required"`
	ItemCode   string `form:"item_code" binding:"required"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
}

type GetBill struct {
	PageRequest
	BillStatus string `form:"bill_status"`
	ServiceId  int64  `form:"service_id" binding:"required"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	BillCode   string `form:"bill_code"`
	ItemCode   string `form:"item_code"`
}

type KioskInventoryInsertBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Code       string `json:"code" binding:"required"` // Mã đơn nhập
	ServiceId  int64  `json:"service_id" binding:"required"`
}

type AddItemToInventoryBody struct {
	ServiceId  int64        `json:"service_id" binding:"required"`
	PartnerUid string       `json:"partner_uid" binding:"required"`
	CourseUid  string       `json:"course_uid" binding:"required"`
	ListItem   ListItemBody `json:"list_item" binding:"required"`
}

type ListItemBody []ItemBody

func (item *ListItemBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListItemBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type ItemBody struct {
	ItemCode    string `json:"item_code" binding:"required"`
	ServiceType string `json:"service_type" binding:"required"`
	Quantity    int64  `json:"quantity" binding:"required"`
}
