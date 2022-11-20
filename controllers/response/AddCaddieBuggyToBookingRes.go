package response

import (
	"start/models"
	model_booking "start/models/booking"
)

// BuggyCalendarResponse : Booking
type AddCaddieBuggyToBookingRes struct {
	Booking   model_booking.Booking
	NewCaddie models.Caddie
	NewBuggy  models.Buggy
	OldCaddie models.Caddie
	OldBuggy  models.Buggy
}
