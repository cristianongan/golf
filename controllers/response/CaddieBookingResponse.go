package response

import (
	"start/models"
	model_booking "start/models/booking"
)

// CaddieBookingResponse : Booking
type CaddieBookingResponse struct {
	models.Model
	PartnerUid   string                       `json:"partner_uid,omitempty"`
	CourseUid    string                       `json:"course_uid,omitempty"`
	BookingDate  string                       `json:"booking_date,omitempty"`
	CaddieId     string                       `json:"caddie_id,omitempty"`
	CaddieInfo   *model_booking.BookingCaddie `json:"caddie_info,omitempty"`
	CustomerName string                       `json:"customer_name,omitempty"`
	CustomerInfo *model_booking.CustomerInfo  `json:"customer_info,omitempty"`
	TeeTime      string                       `json:"tee_time,omitempty"`
	Hole         int                          `json:"hole,omitempty"`
}
