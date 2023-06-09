package utils

import (
	"database/sql/driver"
	"encoding/json"
)

// ------- List Gs Of Guest ---------
type ListGsOfGuest []GsOfGuest

type GsOfGuest struct {
	GuestStyle string `json:"guest_style"`
	Dow        string `json:"dow"`
}

func (item *ListGsOfGuest) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListGsOfGuest) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ------- List Golf Hole ---------
type ListGolfHoleFee []GolfHoleFee

func (item *ListGolfHoleFee) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListGolfHoleFee) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ------- List Subbag ---------
type ListSubBag []BookingSubBag

func (item *ListSubBag) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListSubBag) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ------- List Order Item ---------
type ListOrderItem []OrderItem

type OrderItem struct {
	ItemCode string `json:"item_code"`
	Quantity int    `json:"quantity"`
	Type     string `json:"type"`
}

func (item *ListOrderItem) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListOrderItem) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ------- List Int -------

type ListInt []int

func (item *ListInt) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListInt) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ------- List Int64 -------

type ListInt64 []int64

func (item *ListInt64) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListInt64) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ==================================================
type ListString []string

func (item *ListString) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListString) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

func AddItemToListWithCheckExits(s []string, e string) []string {
	pos := ContainString(s, e)
	if pos >= 0 {
		return s
	}
	return append(s, e)
}

func ContainString(s []string, e string) int {
	for index, a := range s {
		if a == e {
			return index
		}
	}
	return -1
}

func DeleteItemInStringArray(s []string, index int) []string {
	copy(s[index:], s[index+1:]) // Shift s[i+1:] left one index.
	s[len(s)-1] = ""             // Erase last element (write zero value).
	return s[:len(s)-1]          // Truncate slice.
}

type BookingAgencyPayForBagData struct {
	Type string `json:"type"` // GOLF_FEE, BUGGY_FEE, BOOKING_CADDIE_FEE
	Fee  int64  `json:"fee"`
	Name string `json:"name"`
	Hole int    `json:"hole"`
}

type ListBookingAgencyPayForBagData []BookingAgencyPayForBagData

func (item *ListBookingAgencyPayForBagData) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingAgencyPayForBagData) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ------- List Interface ---------
type ListInterface []map[string]interface{}

func (item *ListInterface) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListInterface) Value() (driver.Value, error) {
	return json.Marshal(&item)
}
