package request

type GetBookingServiceItem struct {
	PageRequest
	GroupCode string `form:"group_code"`
	ServiceId string `form:"service_id"`
	Name      string `form:"name"`
}
