package request

type CreateMaintainScheduleBody struct {
	MaintainScheduleList []CreateMaintainSchedule `json:"maintain_schedule_list"`
}

type CreateMaintainSchedule struct {
	CourseName      string                        `json:"course_name"`
	WeekId          string                        `json:"week_id"` // 2023-12`
	ApplyDayOffList []MaintainScheduleApplyDayOff `json:"apply_day_off_list"`
}

type MaintainScheduleApplyDayOff struct {
	MorningOff       *bool  `json:"morning_off"`
	AfternoonOff     *bool  `json:"afternoon_off"`
	MorningTimeOff   string `json:"morning_time_off"`
	AfternoonTimeOff string `json:"afternoon_time_off"`
	ApplyDate        string `json:"apply_date"`
}

type GetMaintainScheduleList struct {
	PageRequest
	WeekId string `form:"week_id"`
}

type UpdateMaintainScheduleBody struct {
	CourseName      string                        `json:"course_name"`
	WeekId          string                        `json:"week_id"` // 2023-12`
	ApplyDayOffList []MaintainScheduleApplyDayOff `json:"apply_day_off_list"`
}
