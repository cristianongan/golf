package request

type AddItemServiceCartBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	GolfBag    string `json:"golf_bag" binding:"required"`
	ItemCode   string `json:"item_code" binding:"required"`
	Quantity   int64  `json:"quantity"`
	ServiceId  int64  `json:"service_id"`
	BillId     int64  `json:"bill_id"`
}

type AddDiscountServiceItemBody struct {
	CartItemId     int64  `json:"cart_item_id"`
	DiscountType   string `json:"discount_type"`
	DiscountPrice  int64  `json:"discount_price"`
	DiscountReason string `json:"discount_reason"`
}

type GetItemServiceCartBody struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	GolfBag     string `form:"golf_bag"`
	BookingDate string `form:"booking_date"`
	ServiceId   int64  `form:"service_id"`
	BillId      int64  `form:"bill_id"`
}

type GetBestItemBody struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	ServiceId  int64  `form:"service_id"`
	GroupCode  string `form:"group_code"`
}

type GetServiceCartBody struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date" binding:"required"`
	ServiceId   int64  `form:"service_id"`
}

type UpdateServiceCartBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	CartItemId int64  `json:"cart_item_id"`
	Quantity   int64  `json:"quantity"`
	Note       string `json:"note"`
}

type CreateBillCodeBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	GolfBag    string `json:"golf_bag"`
	ServiceId  int64  `json:"service_id"`
}

type MoveItemToOtherServiceCartBody struct {
	ServiceCartId  int64   `json:"service_cart_id"`
	GolfBag        string  `json:"golf_bag"`
	CartItemIdList []int64 `json:"cart_item_id_list"`
}

type CreateNewGuestBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	GuestName  string `json:"guest_name"`
}

type FinishOrderBody struct {
	BillId int64 `json:"bill_id" binding:"required"`
}
