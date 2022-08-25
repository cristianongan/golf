package request

import model_booking "start/models/booking"

type AddBookingServiceItem struct {
	BookingUid  string `json:"booking_uid"`  // Cho ở lễ tân
	Bag         string `json:"bag"`          // ở GO thì chỉ cần truyền bag và ngày
	BookingDate string `json:"booking_date"` // ở GO thì chỉ cần truyền bag và ngày
	model_booking.BookingServiceItem
}

type UpdateBookingServiceItem struct {
	model_booking.BookingServiceItem
}
