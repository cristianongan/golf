package request

import (
	"database/sql/driver"
	"encoding/json"
)

type ValetAddListBagCaddieBuggyToBooking struct {
	Data ListBagCaddieBuggy `json:"data"`
}

type AddBagCaddieBuggyToBooking struct {
	PartnerUid     string `json:"partner_uid"`
	CourseUid      string `json:"course_uid"`
	BookingUid     string `json:"booking_uid"`
	Bag            string `json:"bag"`
	CaddieCode     string `json:"caddie_code"`
	BuggyCode      string `json:"buggy_code"`
	BookingDate    string `json:"booking_date"`
	IsPrivateBuggy bool   `json:"is_private_buggy"`
}

type ListBagCaddieBuggy []AddBagCaddieBuggyToBooking

func (item *ListBagCaddieBuggy) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListBagCaddieBuggy) Value() (driver.Value, error) {
	return json.Marshal(&item)
}
