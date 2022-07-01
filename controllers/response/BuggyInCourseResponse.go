package response

import (
	"start/models"
	model_booking "start/models/booking"
)

type BuggyInCourseResponse struct {
	models.Model
	PartnerUid  string                      `json:"partner_uid,omitempty"`
	CourseUid   string                      `json:"course_uid,omitempty"`
	BookingDate string                      `json:"booking_date,omitempty"`
	BuggyId     string                      `json:"buggy_id,omitempty"`
	BuggyInfo   *model_booking.BookingBuggy `json:"buggy_info,omitempty"`
	Bag         string                      `json:"bag,omitempty"`
	CaddieId    string                      `json:"caddie_id,omitempty"`
	CaddieInfo  model_booking.BookingCaddie `json:"caddie_info,omitempty"`
	TeeType     string                      `json:"tee_type,omitempty"`
	TeeOffTime  string                      `json:"tee_off_time,omitempty"`
}
