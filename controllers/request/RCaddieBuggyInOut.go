package request

type GetListCaddieBuggyInOut struct {
	PageRequest
	PartnerUid     string `form:"partner_uid" binding:"required"`
	CourseUid      string `form:"course_uid" binding:"required"`
	BookingDate    string `form:"booking_date"`
	Bag            string `form:"bag"`
	CaddieType     string `form:"caddie_type"`
	BuggyType      string `form:"buggy_type"`
	BuggCode       string `form:"buggy_code"`
	CaddieCode     string `form:"caddie_code"`
	ShareBuggy     *bool  `form:"share_buggy"`
	BagOrBuggyCode string `form:"bag_or_buggy_code"`
}

type RCaddieSlotExample struct {
	Caddie string `form:"caddie"`
}
