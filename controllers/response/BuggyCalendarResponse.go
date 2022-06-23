package response

import (
	"start/models"
	model_booking "start/models/booking"
)

// BuggyCalendarResponse : Booking
type BuggyCalendarResponse struct {
	models.Model
	PartnerUid  string                          `json:"partner_uid,omitempty"`
	CourseUid   string                          `json:"course_uid,omitempty"`
	BookingDate string                          `json:"booking_date,omitempty"`
	BuggyId     string                          `json:"buggy_id,omitempty"`
	BuggyInfo   *model_booking.BookingBuggy     `json:"buggy_info,omitempty"`
	Rounds      *model_booking.ListBookingRound `json:"rounds,omitempty"`
}
