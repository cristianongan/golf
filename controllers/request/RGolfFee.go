package request

type GetListGolfFeeForm struct {
	PageRequest
	PartnerUid       string `form:"partner_uid"`
	CourseUid        string `form:"course_uid"`
	Status           string `form:"status"`
	TablePriceId     int64  `form:"table_price_id"`
	GroupId          int64  `form:"group_id"`
	CustomerType     string `form:"customer_type"`     // GUEST, AGENCY
	CustomerCategory string `form:"customer_category"` // CUSTOMER, AGENCY
	GuestStyle       string `form:"guest_style"`
	GuestStyleName   string `form:"guest_style_name"`
}

type GetListGolfFeeByGuestStyleForm struct {
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	GuestStyle string `form:"guest_style"`
}
