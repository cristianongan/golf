package request

type CreateLockTurn struct {
	PartnerUid     string `json:"partner_uid" binding:"required"`
	CourseUid      string `json:"course_uid" binding:"required"`
	TeeTime        string `json:"tee_time" binding:"required"`
	TurnTimeStatus string `json:"turn_time_status" binding:"required"`
	BookingDate    string `json:"booking_date" binding:"required"`
	Tee            string `json:"tee" binding:"required"`
	Note           string `json:"note"`
}
type GetListLockTurn struct {
	PageRequest
	PartnerUid     string `form:"partner_uid"`
	CourseUid      string `form:"course_uid"`
	BookingDate    string `json:"booking_date"`
	TurnTimeStatus string `form:"turn_time_status"`
}
