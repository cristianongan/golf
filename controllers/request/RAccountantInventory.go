package request

import (
	"database/sql/driver"
	"encoding/json"
)

type AccountantAddInventory struct {
	PartnerUid    string                          `json:"partner_uid" binding:"required"`
	CourseUid     string                          `json:"course_uid" binding:"required"`
	InventoryCode string                          `json:"ma_kho" binding:"required"`
	ListItem      ListAccountantInventoryItemBody `json:"ds_sp" binding:"required"`
	Note          string                          `json:"note"`
	OutputDate    int64                           `json:"ngay_tao"`
}

type AccountantInventoryItemBody struct {
	ItemCode string  `json:"ma_sp"`
	Quantity int64   `json:"so_luong"`
	Unit     string  `json:"dv"`
	Price    float64 `json:"gia_ban"`
}

type ListAccountantInventoryItemBody []AccountantInventoryItemBody

func (item *ListAccountantInventoryItemBody) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListAccountantInventoryItemBody) Value() (driver.Value, error) {
	return json.Marshal(&item)
}
