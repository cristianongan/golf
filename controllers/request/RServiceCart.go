package request

type AddItemServiceCartBody struct {
	GolfBag   string `json:"golf_bag"`
	ItemCode  string `json:"item_code"`
	Quantity  int64  `json:"quantity"`
	ServiceId int64  `json:"service_id"`
	GroupType string `json:"group_type"`
	BillId    int64  `json:"bill_id"`
}

type AddDiscountServiceItemBody struct {
	CartItemId     int64   `json:"cart_item_id"`
	DiscountType   string  `json:"discount_type"`
	DiscountPrice  float64 `json:"discount_price"`
	DiscountReason string  `json:"discount_reason"`
}

type GetItemServiceCartBody struct {
	PageRequest
	GolfBag     string `json:"golf_bag"`
	BookingDate string `json:"booking_date"`
	ServiceId   int64  `json:"service_id"`
	BillId      int64  `json:"bill_id"`
}

type GetBestItemBody struct {
	PageRequest
	ServiceId int64  `json:"service_id"`
	GroupCode string `json:"group_code"`
}

type GetServiceCartBody struct {
	PageRequest
	BookingDate string `json:"booking_date" binding:"required"`
	ServiceId   int64  `json:"service_id"`
}

type UpdateServiceCartBody struct {
	CartItemId int64  `json:"cart_item_id"`
	Quantity   int64  `json:"quantity"`
	Note       string `json:"note"`
}

type CreateBillCodeBody struct {
	GolfBag   string `json:"golf_bag"`
	ServiceId int64  `json:"service_id"`
}

type MoveItemToOtherServiceCartBody struct {
	ServiceCartId  int64   `json:"service_cart_id"`
	GolfBag        string  `json:"golf_bag"`
	CartItemIdList []int64 `json:"cart_item_id_list"`
}
