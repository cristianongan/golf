package response

import (
	"database/sql/driver"
	"encoding/json"
)

// GolfBagResponse : Booking
type ReportSubBagResponse struct {
	Uid        string `json:"uid"`
	PartnerUid string `json:"partner_uid"` // Hang Golf
	CourseUid  string `json:"course_uid"`  // San Golf

	BookingDate  string `json:"booking_date"`   // Ex: 06/11/2022
	CheckOutTime int64  `json:"check_out_time"` // Time Check Out
	BagStatus    string `json:"bag_status"`     // Bag status

	Bag string `json:"bag"` // Golf Bag

	MyCost   int64 `json:"my_cost"`
	ToBePaid int64 `json:"to_be_paid"`
}
type ListReportSubBagResponse []ReportSubBagResponse

func (item *ListReportSubBagResponse) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), item)
}

func (item ListReportSubBagResponse) Value() (driver.Value, error) {
	return json.Marshal(&item)
}

type ReportMainBagResponse struct {
	ReportSubBagResponse
	SubBag ListReportSubBagResponse `json:"sub_bag"`
}
