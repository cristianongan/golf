package request

type GetBuggyCaddyFeeSetting struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Hole       int    `form:"hole" binding:"required"`
	GuestStyle string `form:"guest_style"`
	AgencyId   int64  `form:"agency_id"`
}
