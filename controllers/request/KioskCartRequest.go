package request

type AddItemToKioskCartBody struct {
	GolfBag string `json:"golf_bag"`
	//BookingDate string `json:"booking_date"`
	ItemCode  string `json:"item_code"`
	Quantity  int64  `json:"quantity"`
	KioskCode string `json:"kiosk_code"`
}

type AddDiscountToKioskItemBody struct {
	CartItemId     int64   `json:"cart_item_id"`
	DiscountType   string  `json:"discount_type"`
	DiscountPrice  float64 `json:"discount_price"`
	DiscountReason string  `json:"discount_reason"`
}

type GetItemInKioskCartBody struct {
	PageRequest
	GolfBag     string `form:"golf_bag"`
	BookingDate string `form:"booking_date"`
	KioskCode   string `json:"kiosk_code"`
}

type UpdateQuantityToKioskCartBody struct {
	CartItemId int64 `json:"cart_item_id"`
	Quantity   int64 `json:"quantity"`
}

type DeleteItemInKioskCartBody struct {
	CartItemId int64 `json:"cart_item_id"`
}

type CreateKioskBillingBody struct {
	GolfBag     string `json:"golf_bag"`
	BookingDate string `json:"booking_date"`
	KioskCode   string `json:"kiosk_code"`
}

type MoveItemToOtherKioskCartBody struct {
	CartCode       string  `json:"cart_code"`
	GolfBag        string  `json:"golf_bag"`
	CartItemIdList []int64 `json:"cart_item_id_list"`
}
