package request

type GetDetalCaddieWorkingSyncBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Week       int    `json:"week" binding:"required"`
	EmployeeId string `json:"employee_id" binding:"required"`
}

type CreateCaddieWorkingReq struct {
	PartnerUid        string               `json:"partner_uid"`
	CourseUid         string               `json:"course_uid"`
	CaddieWorkingList []CaddieCalendarList `json:"caddie_calendar_list" binding:"required"`
}

type CaddieWorkingList struct {
	EmployeeID string `json:"employee_id"`
	TimeStart  string `json:"time_start"`
	TimeEnd    string `json:"time_end"`
	ApplyDate  string `json:"apply_date"`
}

type CaddieCalendarList struct {
	ApplyMonth string              `json:"apply_month"`
	CaddieList []CaddieWorkingList `json:"caddie_list"`
}
