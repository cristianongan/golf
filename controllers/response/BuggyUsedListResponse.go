package response

import (
	"start/models"
	model_booking "start/models/booking"
	"start/utils"
)

// BuggyUsedListResponse : Booking
type BuggyUsedListResponse struct {
	models.Model
	PartnerUid     string                       `json:"partner_uid,omitempty"`
	CourseUid      string                       `json:"course_uid,omitempty"`
	BookingDate    string                       `json:"booking_date,omitempty"`
	BuggyId        string                       `json:"buggy_id,omitempty"`
	BuggyInfo      *model_booking.BookingBuggy  `json:"buggy_info,omitempty"`
	CaddieId       string                       `json:"caddie_id,omitempty"`
	CaddieInfo     *model_booking.BookingCaddie `json:"caddie_info,omitempty"`
	Bag            string                       `json:"bag,omitempty"`
	CustomerName   string                       `json:"customer_name,omitempty"`
	CustomerInfo   *model_booking.CustomerInfo  `json:"customer_info,omitempty"`
	Hole           int                          `json:"hole,omitempty"`
	FlightId       int64                        `json:"flight_id,omitempty"`
	AgencyId       int64                        `json:"agency_id,omitempty"`
	AgencyInfo     *model_booking.BookingAgency `json:"agency_info,omitempty"`
	MainBags       *utils.ListSubBag            `json:"main_bags,omitempty"`
	SubBags        *utils.ListSubBag            `json:"sub_bags,omitempty"`
	IsPrivateBuggy *bool                        `json:"is_private_buggy"`
	MemberCardUid  string                       `json:"member_card_uid"`
}
