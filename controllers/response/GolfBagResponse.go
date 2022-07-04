package response

import (
	"start/models"
	model_booking "start/models/booking"
)

// GolfBagResponse : Booking
type GolfBagResponse struct {
	models.Model
	PartnerUid  string                      `json:"partner_uid,omitempty"`
	CourseUid   string                      `json:"course_uid,omitempty"`
	BookingDate string                      `json:"booking_date,omitempty"`
	Bag         string                      `json:"bag,omitempty"`
	BagStatus   string                      `json:"bag_status,omitempty"`
	CaddieId    string                      `json:"caddie_id,omitempty"`
	CaddieInfo  model_booking.BookingCaddie `json:"caddie_info,omitempty"`
	BuggyId     string                      `json:"buggy_id,omitempty"`
	BuggyInfo   *model_booking.BookingBuggy `json:"buggy_info,omitempty"`
	FlightId    int64                       `json:"flight_id"`
}
