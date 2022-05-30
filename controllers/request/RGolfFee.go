package request

type GetListGolfFeeForm struct {
	PageRequest
	PartnerUid   string `form:"partner_uid"`
	CourseUid    string `form:"course_uid"`
	Status       string `form:"status"`
	TablePriceId int64  `form:"table_price_id"`
	GroupId      int64  `form:"group_id"`
}
