package request

type CreateCaddieWorkingCalendarBody struct {
	CaddieList []CaddieWorkingCalendarRequest `json:"caddie_list" binding:"required"`
}

type GetCaddieWorkingCalendarList struct {
	PartnerUid string `form:"partner_uid" validate:"required"`
	CourseUid  string `form:"course_uid" validate:"required"`
	ApplyDate  string `form:"apply_date" validate:"required"`
}

type UpdateCaddieWorkingCalendarBody struct {
	CaddieCode string `json:"caddie_code" validate:"required"`
}

type CaddieWorkingCalendarRequest struct {
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	CaddieCode  string `json:"caddie_code"`
	ApplyDate   string `json:"apply_date"`
	NumberOrder int64  `json:"number_order"`
}
