package request

type GetBuggyCalendar struct {
	PageRequest
	BuggyCode string `form:"buggy_code"`
	Month     string `form:"month"`
}
