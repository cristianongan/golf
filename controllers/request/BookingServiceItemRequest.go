package request

type GetBookingServiceItem struct {
	PageRequest
	GroupCode string `form:"group_code"`
}
