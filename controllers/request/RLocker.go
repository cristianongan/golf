package request

type GetListLockerForm struct {
	PageRequest
	PartnerUid   string `form:"partner_uid" binding:"required"`
	CourseUid    string `form:"course_uid"`
	Locker       string `form:"locker"`
	GolfBag      string `form:"golf_bag"`
	LockerStatus string `form:"locker_status"`
	From         int64  `form:"from"`
	To           int64  `form:"to"`
}

type ReturnLockerReq struct {
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid" binding:"required"`
	BookingDate string `json:"booking_date" binding:"required"`
	LockerNo    string `json:"locker_no" binding:"required"`
}
