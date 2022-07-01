package response

import (
	"start/models"
	model_booking "start/models/booking"
)

type CaddieAgencyBookingResponse struct {
	models.Model
	PartnerUid  string                       `json:"partner_uid,omitempty"`
	CourseUid   string                       `json:"course_uid,omitempty"`
	BookingDate string                       `json:"booking_date,omitempty"`
	TeeTime     string                       `json:"tee_time,omitempty"`
	Hole        int                          `json:"hole,omitempty"`
	AgencyId    int64                        `json:"agency_id,omitempty"`
	AgencyInfo  *model_booking.BookingAgency `json:"agency_info,omitempty"`
}
