package request

type AddItemServiceCartBody struct {
	GolfBag   string `form:"golf_bag"`
	ItemCode  string `form:"item_code"`
	Quantity  int64  `form:"quantity"`
	ServiceId int64  `form:"service_id"`
	GroupType string `form:"group_type"`
}

type AddDiscountServiceItemBody struct {
	CartItemId     int64   `form:"cart_item_id"`
	DiscountType   string  `form:"discount_type"`
	DiscountPrice  float64 `form:"discount_price"`
	DiscountReason string  `form:"discount_reason"`
}

type GetItemServiceCartBody struct {
	PageRequest
	GolfBag     string `form:"golf_bag"`
	BookingDate string `form:"booking_date"`
	ServiceId   int64  `form:"service_id"`
	BillId      int64  `form:"bill_id"`
}

type GetBestItemBody struct {
	PageRequest
	ServiceId int64  `form:"service_id"`
	GroupCode string `form:"group_code"`
}

type GetServiceCartBody struct {
	PageRequest
	BookingDate string `form:"booking_date" binding:"required"`
	ServiceId   int64  `form:"service_id"`
}

type UpdateServiceCartBody struct {
	CartItemId int64  `form:"cart_item_id"`
	Quantity   int64  `form:"quantity"`
	Note       string `form:"note"`
}

type CreateBillCodeBody struct {
	GolfBag   string `form:"golf_bag"`
	ServiceId int64  `form:"service_id"`
}

type MoveItemToOtherServiceCartBody struct {
	ServiceCartId  int64   `form:"service_cart_id"`
	GolfBag        string  `form:"golf_bag"`
	CartItemIdList []int64 `form:"cart_item_id_list"`
}
