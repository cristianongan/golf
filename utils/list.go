package utils

import (
	"database/sql/driver"
	"encoding/json"
)

// ------- Booking Service --------
type BookingService struct {
	Order         string `json:"order"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Quality       int    `json:"quality"`
	UnitPrice     int64  `json:"unit_price"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	Amount        int64  `json:"amount"`
	Input         string `json:"input"`
}

type ListBookingServices []BookingService

func (item *ListBookingServices) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingServices) Value() (driver.Value, error) {
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
