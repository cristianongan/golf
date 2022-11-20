package request

type AddItemToKioskCartBody struct {
	GolfBag string `json:"golf_bag"`
	//BookingDate string `json:"booking_date"`
	ItemCode  string `json:"item_code"`
	Quantity  int64  `json:"quantity"`
	KioskCode int64  `json:"kiosk_code"`
	KioskType string `json:"kiosk_type"`
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
	KioskCode   int64  `form:"kiosk_code"`
}

type GetBestItemInKioskBody struct {
	PageRequest
	KioskCode int64  `form:"kiosk_code"`
	GroupCode string `form:"group_code"`
}

type GetCartInKioskBody struct {
	PageRequest
	BookingDate string `form:"booking_date" binding:"required"`
	KioskCode   int64  `form:"kiosk_code" binding:"required"`
}

type UpdateQuantityToKioskCartBody struct {
	CartItemId int64  `json:"cart_item_id"`
	Quantity   int64  `json:"quantity"`
	Note       string `json:"note"`
}

type DeleteItemInKioskCartBody struct {
	CartItemId int64 `json:"cart_item_id"`
}

type CreateKioskBillingBody struct {
	GolfBag string `json:"golf_bag"`
	//BookingDate string `json:"booking_date"`
	KioskCode int64 `json:"kiosk_code"`
}

type MoveItemToOtherKioskCartBody struct {
	CartCode       string  `json:"cart_code"`
	GolfBag        string  `json:"golf_bag"`
	CartItemIdList []int64 `json:"cart_item_id_list"`
}
