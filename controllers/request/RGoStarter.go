package request

type GetBookingForCaddieOnCourseForm struct {
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"`
	BookingDate string `form:"booking_date" binding:"required"`
	Bag         string `form:"bag"`
	PlayerName  string `form:"player_name"`
	BuggyId     int64  `form:"buggy_id"`
	CaddieId    int64  `form:"caddie_id"`
	InFlight    string `form:"in_flight"`
}

// Add Caddie, Buggy To Booking
type AddCaddieBuggyToBooking struct {
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	Bag         string `json:"bag"`
	CaddieCode  string `json:"caddie_code"`
	BuggyCode   string `json:"buggy_code"`
	BookingDate string `json:"booking_date"`
}

type CreateFlightBody struct {
	PartnerUid  string                 `json:"partner_uid"`
	CourseUid   string                 `json:"course_uid"`
	BookingDate string                 `json:"booking_date"`
	ListData    []CaddieBuggyToBooking `json:"list_data"`
	Note        string                 `json:"note"`
	Tee         int                    `json:"tee"`     // Tee
	TeeOff      string                 `json:"tee_off"` // Tee Off
	CourseType  string                 `json:"course_type"`
}

type CaddieBuggyToBooking struct {
	Bag            string `json:"bag"`
	CaddieCode     string `json:"caddie_code"`
	BuggyCode      string `json:"buggy_code"`
	IsPrivateBuggy bool   `json:"is_private_buggy"`
}

type OutCaddieBody struct {
	BookingUid  string `json:"booking_uid" binding:"required"`
	CaddieHoles int    `json:"caddie_holes"`
	GuestHoles  int    `json:"guest_holes"`
	Note        string `json:"note"`
}

type DeleteAttachBody struct {
	BookingUid  string `json:"booking_uid" binding:"required"`
	IsOutCaddie *bool  `json:"is_out_caddie"`
	IsOutBuggy  *bool  `json:"is_out_buggy"`
	Note        string `json:"note"`
}

type OutAllFlightBody struct {
	FlightId    int64  `json:"flight_id" binding:"required"`
	CaddieHoles int    `json:"caddie_holes"`
	GuestHoles  int    `json:"guest_holes"`
	Note        string `json:"note"`
}

type SimpleOutFlightBody struct {
	FlightId    int64  `json:"flight_id" binding:"required"`
	Bag         string `json:"bag" binding:"required"`
	CaddieHoles int    `json:"caddie_holes"`
	GuestHoles  int    `json:"guest_holes"`
	Note        string `json:"note"`
}

type NeedMoreCaddieBody struct {
	BookingUid  string `json:"booking_uid" binding:"required"`
	CaddieCode  string `json:"new_caddie_code" binding:"required"`
	CaddieHoles int    `json:"old_caddie_holes"`
	Note        string `json:"note"`
}

type GetStartingSheetForm struct {
	PageRequest
	PartnerUid           string `form:"partner_uid" binding:"required"`
	CourseUid            string `form:"course_uid" binding:"required"`
	BookingDate          string `form:"booking_date"`
	Bag                  string `form:"bag"`
	CaddieCode           string `form:"caddie_code"`
	CaddieName           string `form:"caddie_name"`
	CustomerName         string `form:"customer_name"`
	NumberPeopleInFlight *int64 `form:"number_people"`
}

type ChangeCaddieBody struct {
	BookingUid string `json:"booking_uid"`
	CaddieCode string `json:"caddie_code"`
	Reason     string `json:"reason"`
	Note       string `json:"note"`
}

type ChangeBuggyBody struct {
	BookingUid     string `json:"booking_uid"`
	BuggyCode      string `json:"buggy_code"`
	Reason         string `json:"reason"`
	Note           string `json:"note"`
	IsPrivateBuggy bool   `json:"is_private_buggy"`
}

type EditHolesOfCaddiesBody struct {
	BookingUid string `json:"booking_uid"`
	CaddieCode string `json:"caddie_code"`
	Hole       int    `json:"hole"`
}

//type AddBagToFlightBody struct {
//	BookingUid string `json:"booking_uid"`
//	GolfBag    string `json:"golf_bag"`
//	FlightId   int64  `json:"flight_id"`
//	CaddieCode string `json:"caddie_code"`
//	BuggyCode  string `json:"buggy_code"`
//}

type AddBagToFlightBody struct {
	BookingDate string                 `json:"booking_date"`
	FlightId    int64                  `json:"flight_id"`
	ListData    []CaddieBuggyToBooking `json:"list_data"`
}

type GetFlightList struct {
	PageRequest
	BookingDate          string `form:"booking_date"`
	PeopleNumberInFlight *int   `form:"people_number_in_flight"`
	PartnerUid           string `form:"partner_uid"`
	CourseUid            string `form:"course_uid"`
	GolfBag              string `form:"bag"`
	CaddieName           string `form:"caddie_name"`
	CustomerName         string `form:"customer_name"`
	CaddieCode           string `form:"caddie_code"`
}

type MoveBagToFlightBody struct {
	BookingUid string `json:"booking_uid"`
	GolfBag    string `json:"golf_bag"`
	FlightId   int64  `json:"flight_id"`
	HolePlayed int64  `json:"hole_played"`
}

type CheckoutBody struct {
	BookingUid string `json:"booking_uid" validate:"required"`
	GolfBag    string `json:"golf_bag" validate:"required"`
}
