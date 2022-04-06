package utils

import (
	"database/sql/driver"
	"encoding/json"
)

// ------- List Subbag ---------
type ListSubBag []BookingSubBag

func (item *ListSubBag) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListSubBag) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

// ------- List Booking service ---------
type ListBookingServiceItems []BookingServiceItem

func (item *ListBookingServiceItems) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBookingServiceItems) Value() (driver.Value, error) {
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
