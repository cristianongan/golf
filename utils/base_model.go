package utils

// -------- Booking Sub Bag ------
type BookingSubBag struct {
	GolfBag    string `json:"golf_bag"`
	BookingUid string `json:"booking_uid"`
}

// ------- Booking Service item --------
type BookingServiceItem struct {
	Type          string `json:"type"`  // Loại rental, kiosk, proshop,...
	Order         string `json:"order"` // Có thể là mã
	Name          string `json:"name"`
	Code          string `json:"code"`
	Quality       int    `json:"quality"` // Số lượng
	UnitPrice     int64  `json:"unit_price"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	Amount        int64  `json:"amount"`
	Input         string `json:"input"`
}
