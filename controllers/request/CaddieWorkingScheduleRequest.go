package request

type CreateWorkingScheduleBody struct {
	CaddieGroupList []CreateWorkingScheduleForOneGroup `json:"caddie_group_list"`
}

type CreateWorkingScheduleForOneGroup struct {
	CaddieGroupCode string        `json:"caddie_group_code"`
	WeekId          string        `json:"week_id"` // 2023-12`
	ApplyDayOffList []ApplyDayOff `json:"apply_day_off_list"`
}

type ApplyDayOff struct {
	ApplyDate string `json:"apply_date"`
	IsDayOff  bool   `json:"is_day_off"`
}

type GetCaddieWorkingScheduleList struct {
	PageRequest
	WeekId string `form:"week_id"`
}

type UpdateWorkingScheduleBody struct {
	CaddieGroupCode string        `json:"caddie_group_code"`
	WeekId          string        `json:"week_id"` // 2023-12`
	ApplyDayOffList []ApplyDayOff `json:"apply_day_off_list"`
}
