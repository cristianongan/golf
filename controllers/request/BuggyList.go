package request

type GetBuggyList struct {
	PageRequest
	BuggyCode  string `form:"buggy_code"`
	FromDate   string `form:"from_date"`
	ToDate     string `form:"to_date"`
	CaddieCode string `form:"caddie_code"`
	GolfBag    string `form:"golf_bag"`
	IsTimeOut  string `form:"is_time_out"`
}
