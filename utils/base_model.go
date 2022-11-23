package utils

import (
	"database/sql/driver"
	"encoding/json"
)

// -------- Booking Sub Bag ------
type BookingSubBag struct {
	GolfBag     string                         `json:"golf_bag"` // Có thể bỏ
	BookingUid  string                         `json:"booking_uid"`
	PlayerName  string                         `json:"player_name"`
	BillCode    string                         `json:"bill_code"`
	BookingCode string                         `json:"booking_code"`
	CmsUser     string                         `json:"cms_user"`
	CmsUserLog  string                         `json:"cms_user_log"`
	AgencyPaid  ListBookingAgencyPayForBagData `json:"sub_agency_paid,omitempty"`
}

type GolfHoleFee struct {
	Hole int   `json:"hole"`
	Fee  int64 `json:"fee"`
}

type CountStruct struct {
	Count int64 `json:"count"`
}

type TotalStruct struct {
	TotalAmount int64 `json:"total_amount"`
}

type CountAnnualFeeStruct struct {
	TotalA int64 `json:"total_a"`
	TotalB int64 `json:"total_b"`
	TotalC int64 `json:"total_c"`
	TotalD int64 `json:"total_d"`
	TotalE int64 `json:"total_e"`
	TotalG int64 `json:"total_g"`
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
