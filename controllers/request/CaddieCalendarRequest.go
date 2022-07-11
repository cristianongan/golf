package request

type CreateCaddieCalendarBody struct {
	CaddieUidList []int64 `json:"caddie_uid_list" validate:"required"`
	Title         string  `json:"title" validate:"required"`
	DayOffType    string  `json:"day_off_type" validate:"required"`
	FromDate      string  `json:"from_date" validate:"required"`
	ToDate        string  `json:"to_date" validate:"required"`
	Note          string  `json:"note"`
}

type GetCaddieCalendarList struct {
	PageRequest
	CaddieName string `form:"caddie_name"`
	CaddieCode string `form:"caddie_code"`
	Month      string `form:"month" validate:"required"`
}

type UpdateCaddieCalendarBody struct {
	CaddieUid  string `json:"caddie_uid" validate:"required"`
	Title      string `json:"title"`
	DayOffType string `json:"day_off_type"`
	ApplyDate  string `json:"apply_date,datetime"`
	Note       string `json:"note"`
}

type DeleteCaddieCalendarBody struct {
	CaddieUidList []int64 `json:"caddie_uid_list" validate:"required"`
	Month         string  `json:"month" validate:"required"`
}
