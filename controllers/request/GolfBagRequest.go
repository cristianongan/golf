package request

type GetGolfBagRequest struct {
	PageRequest
	BagStatus   string `form:"bag_status"`
	IsFlight    string `form:"is_flight"`
	BookingDate string `form:"booking_date"`
}
