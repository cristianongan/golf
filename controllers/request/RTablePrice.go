package request

type GetListTablePriceForm struct {
	PageRequest
	PartnerUid     string `form:"partner_uid"`
	CourseUid      string `form:"course_uid"`
	Year           int    `form:"year"`
	TablePriceName string `form:"table_price_name"`
}

type CreateTablePriceBody struct {
	Name       string `json:"name" binding:"required"`
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Status     string `json:"status"`
	FromDate   int64  `json:"from_date"`
	OldPriceId int64  `json:"old_price_id"`
}
