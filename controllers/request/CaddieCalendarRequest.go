package request

type CreateCaddieCalendarBody struct {
	CaddieUid  string `json:"caddie_uid" validate:"required"`
	Title      string `json:"title" validate:"required"`
	DayOffType string `json:"day_off_type" validate:"required"`
	ApplyDate  string `json:"apply_date" validate:"required,datetime"`
	Note       string `json:"note"`
}

type GetCaddieCalendarList struct {
	PageRequest
	CourseUid  string `form:"course_uid"`
	CaddieName string `form:"caddie_name"`
	CaddieCode string `form:"caddie_code"`
	Month      string `form:"month"`
}

type UpdateCaddieCalendar struct {
	CaddieUid  string `json:"caddie_uid" validate:"required"`
	Title      string `json:"title"`
	DayOffType string `json:"day_off_type"`
	ApplyDate  string `json:"apply_date,datetime"`
	Note       string `json:"note"`
}
