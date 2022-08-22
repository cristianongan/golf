package request

type CreateLockTurn struct {
	PartnerUid    string `json:"partner_uid" binding:"required"`
	CourseUid     string `json:"course_uid" binding:"required"`
	TeeTime       string `json:"tee_time" binding:"required"`
	TeeTimeStatus string `json:"tee_time_status" binding:"required"`
	BookingDate   string `json:"booking_date" binding:"required"`
	Note          string `json:"note"`
}
type GetListLockTurn struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	TeeTime    string `form:"tee_time"`
}
