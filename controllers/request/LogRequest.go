package request

type GetLogList struct {
	PageRequest
	Category string `form:"category"`
	Code     string `form:"code"`
	Action   string `form:"action"`
}
