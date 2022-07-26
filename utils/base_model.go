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

type GolfHoleFee struct {
	Hole int   `json:"hole"`
	Fee  int64 `json:"fee"`
}

type CountStruct struct {
	Count int64 `json:"count"`
}

type BookingRestaurant struct {
	Enable       bool  `json:"enable"`
	NumberPeople int64 `json:"number_people"`
}

type BookingRental struct {
	Enable        bool  `json:"enable"`
	GolfSetNumber int64 `json:"golf_set_number"`
	BuggyNumber   int64 `json:"buggy_number"`
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

func (item *BookingRestaurant) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingRestaurant) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func (item *BookingRental) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item BookingRental) Value() (driver.Value, error) {
	return json.Marshal(&item)
}
