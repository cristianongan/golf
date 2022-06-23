package request

type CreateCaddieWorkingCalendarBody struct {
	CaddieUid    string `json:"caddie_uid" validate:"required"`
	CaddieCode   string `json:"caddie_code" validate:"required"`
	CaddieLabel  string `json:"caddie_label" validate:"required"`
	CaddieColumn string `json:"caddie_column" validate:"required"`
	CaddieRow    string `json:"caddie_row" validate:"required"`
	RowTime      string `json:"row_time" validate:"required"`
	ApplyDate    string `json:"apply_date" validate:"required,datetime"`
}

type GetCaddieWorkingCalendarList struct {
	PageRequest
	ApplyDate string `json:"apply_date"`
}

type UpdateCaddieWorkingCalendar struct {
	CaddieColumn string `json:"caddie_column" validate:"required"`
	CaddieRow    string `json:"caddie_row" validate:"required"`
	RowTime      string `json:"row_time" validate:"required"`
	ApplyDate    string `json:"apply_date" validate:"required,datetime"`
	CaddieUid    string `json:"caddie_uid" validate:"required"`
	CaddieCode   string `json:"caddie_code" validate:"required"`
	CaddieLabel  string `json:"caddie_label" validate:"required"`
}
