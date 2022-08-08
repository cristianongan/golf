package request

type GetListLockerForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid"`
	Locker     string `form:"locker"`
	GolfBag    string `form:"golf_bag"`
	From       int64  `form:"from"`
	To         int64  `form:"to"`
}
