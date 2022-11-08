package request

type CreateCaddieWorkingCalendarBody struct {
	PartnerUid        string                             `json:"partner_uid"`
	CourseUid         string                             `json:"course_uid"`
	CaddieWorkingList []CaddieWorkingCalendarListRequest `json:"caddie_working_list" binding:"required"`
}

type GetCaddieWorkingCalendarList struct {
	PartnerUid string `form:"partner_uid" validate:"required"`
	CourseUid  string `form:"course_uid" validate:"required"`
	ApplyDate  string `form:"apply_date" validate:"required"`
}

type UpdateCaddieWorkingCalendarBody struct {
	CaddieCode string `json:"caddie_code" validate:"required"`
}

type CaddieWorkingCalendarListRequest struct {
	ApplyDate  string                         `json:"apply_date"`
	Note       string                         `json:"note"`
	CaddieList []CaddieWorkingCalendarRequest `json:"caddie_list"`
}

type CaddieWorkingCalendarRequest struct {
	CaddieCode     string `json:"caddie_code"`
	Row            string `json:"row"`
	NumberOrder    int64  `json:"number_order"`
	CaddieIncrease bool   `json:"caddie_increase"`
}
