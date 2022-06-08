package utils

import (
	"database/sql/driver"
	"encoding/json"
)

// -------- Booking Sub Bag ------
type BookingSubBag struct {
	GolfBag    string `json:"golf_bag"` // Có thể bỏ
	BookingUid string `json:"booking_uid"`
	PlayerName string `json:"player_name"`
}

// ------- Booking Service item --------
type BookingServiceItem struct {
	ItemId        int64  `json:"item_id"`     // Id item
	BookingUid    string `json:"booking_uid"` // Uid booking
	PlayerName    string `json:"player_name"` // Tên người chơi
	Bag           string `json:"bag"`         // Golf Bag
	Type          string `json:"type"`        // Loại rental, kiosk, proshop,...
	Order         string `json:"order"`       // Có thể là mã
	Name          string `json:"name"`
	Code          string `json:"code"`
	Quality       int    `json:"quality"` // Số lượng
	UnitPrice     int64  `json:"unit_price"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	Amount        int64  `json:"amount"`
	Input         string `json:"input"` // Note
}

type GolfHoleFee struct {
	Hole int   `json:"hole"`
	Fee  int64 `json:"fee"`
}

type CountStruct struct {
	Count int64 `json:"count"`
}

// Other Fee
type ListOtherPaid []OtherPaidBody

func (item *ListOtherPaid) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListOtherPaid) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type OtherPaidBody struct {
	Reason string `json:"reason"`
	Amount int64  `json:"amount"`
}
