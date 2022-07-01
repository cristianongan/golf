package request

type GetBuggyUsedList struct {
	PageRequest
	BuggyCode string `form:"buggy_code"`
	FromDate  string `form:"from_date"`
	ToDate    string `form:"to_date"`
}
