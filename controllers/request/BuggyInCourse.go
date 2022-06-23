package request

type GetBuggyInCourse struct {
	PageRequest
	BuggyCode  string `form:"buggy_code"`
	CaddieCode string `form:"caddie_code"`
	GolfBag    string `form:"golf_bag"`
}
