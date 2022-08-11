package request

type GetListCaddieFee struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	Month      string `form:"month" binding:"required"`
	CaddieName string `json:"Caddie_Name"`
	CaddieCode string `json:"Caddie_Code"`
}

type GetDetailListCaddieFee struct {
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
	CaddieCode string `form:"caddie_code"`
	Month      string `form:"month"`
}
