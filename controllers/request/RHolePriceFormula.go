package request

type GetListHolePriceFormulaForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}

type CreateHolePriceFormulaBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Hole       int    `json:"hole" binding:"required"` // Hố Golf
	StopByRain string `json:"stop_by_rain"`            // Dừng bởi trời mưa
	StopBySelf string `json:"stop_by_self"`            // Dừng bởi người chơi
}
