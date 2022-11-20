package request

type CreateLockTurn struct {
	PartnerUid  string
	CourseUid   string
	TeeTime     string
	BookingDate string
	TeeType     string
}
type GetListLockTurn struct {
	PageRequest
	PartnerUid     string `form:"partner_uid"`
	CourseUid      string `form:"course_uid"`
	BookingDate    string `json:"booking_date"`
	TurnTimeStatus string `form:"turn_time_status"`
}
