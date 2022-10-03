package request

type GetBookingServiceItem struct {
	PageRequest
	GroupCode string `form:"group_code"`
	ServiceId string `form:"service_id"`
	Name      string `form:"name"`
	Type      string `form:"type"`
	ItemCode  string `form:"item_code"`
	FromDate  string `form:"from_date"`
	ToDate    string `form:"to_date"`
}
