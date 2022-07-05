package response

import (
	"start/models"
	model_booking "start/models/booking"
)

type FlightResponse struct {
	models.Model
	PartnerUid  string                      `json:"partner_uid,omitempty"`
	CourseUid   string                      `json:"course_uid,omitempty"`
	BookingDate string                      `json:"booking_date,omitempty"`
	Bag         string                      `json:"bag,omitempty"`
	CaddieId    string                      `json:"caddie_id,omitempty"`
	CaddieInfo  model_booking.BookingCaddie `json:"caddie_info,omitempty"`
	TeeType     string                      `json:"tee_type,omitempty"`
	BagStatus   string                      `json:"bag_status,omitempty"`
}
