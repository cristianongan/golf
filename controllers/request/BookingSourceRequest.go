package request

import "start/utils"

type UpdateBookingSource struct {
	BookingSourceName string           `json:"booking_source_name"`
	TeeTime           utils.ListString `json:"tee_time" gorm:"type:json"`
	NormalDay         *bool            `json:"normal_day"`
	Weekend           *bool            `json:"week_end"`
	NumberOfDays      int64            `json:"number_of_days"`
	Status            string           `form:"status"`
}
type GetListBookingSource struct {
	PageRequest
	BookingSourceName string `form:"booking_source_name"`
	Status            string `form:"status"`
}
