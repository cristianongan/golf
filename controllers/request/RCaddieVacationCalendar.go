package request

type CreateCaddieVacationCalendarBody struct {
	PartnerUid   string `json:"partner_uid" validate:"required"`
	CourseUid    string `json:"course_uid" validate:"required"`
	CaddieId     int64  `json:"caddie_id" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Color        string `json:"color" validate:"required"`
	DateFrom     int64  `json:"date_from" validate:"required"`
	DateTo       int64  `json:"date_to" validate:"required"`
	NumberDayOff int    `json:"number_day_off"`
	Note         string `json:"note"`
}

type GetCaddieVacationCalendarList struct {
	PageRequest
	PartnerUid string `form:"partner_uid" validate:"required"`
	CourseUid  string `form:"course_uid" validate:"required"`
	CaddieName string `form:"caddie_name"`
	CaddieCode string `form:"caddie_code"`
	Month      int    `form:"month" validate:"required"`
}

type UpdateCaddieVacationCalendarBody struct {
	Title        string `json:"title" validate:"required"`
	Color        string `json:"color" validate:"required"`
	DateFrom     int64  `json:"date_from" validate:"required"`
	DateTo       int64  `json:"date_to" validate:"required"`
	NumberDayOff int    `json:"number_day_off"`
	Note         string `json:"note"`
}
