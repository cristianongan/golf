package response

import (
	"start/models"
	model_booking "start/models/booking"
)

// BuggyListResponse : Booking
type BuggyListResponse struct {
	models.Model
	PartnerUid   string                       `json:"partner_uid,omitempty"`
	CourseUid    string                       `json:"course_uid,omitempty"`
	BookingDate  string                       `json:"booking_date,omitempty"`
	BuggyId      string                       `json:"buggy_id,omitempty"`
	BuggyInfo    *model_booking.BookingBuggy  `json:"buggy_info,omitempty"`
	CaddieId     string                       `json:"caddie_id,omitempty"`
	CaddieInfo   *model_booking.BookingCaddie `json:"caddie_info,omitempty"`
	Bag          string                       `json:"bag,omitempty"`
	CustomerName string                       `json:"customer_name,omitempty"`
	CustomerInfo *model_booking.CustomerInfo  `json:"customer_info,omitempty"`
	Hole         int                          `json:"hole,omitempty"`
	FlightId     int64                        `json:"flight_id,omitempty"`
}
