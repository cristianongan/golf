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

type AddItemRentalCartBody struct {
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	GolfBag     string `json:"golf_bag" binding:"required"`
	ItemCode    string `json:"item_code"`
	Quantity    int64  `json:"quantity"`
	ServiceId   int64  `json:"service_id"`
	BillId      int64  `json:"bill_id"`
	Name        string `json:"name"`
	Price       int64  `json:"price"`
	Hole        int    `json:"hole"`
	CaddieCode  string `json:"caddie_code"`
	ServiceType string `json:"service_type"`
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
	BillStatus  string `form:"bill_status"`
}

type GetBestItemBody struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	ServiceId  int64  `form:"service_id"`
	GroupCode  string `form:"group_code"`
}

type GetBestGroupBody struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	ServiceId  string `form:"service_id"`
}

type GetServiceCartBody struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	BookingDate string `form:"booking_date" binding:"required"`
	ServiceId   int64  `form:"service_id"`
}

type GetServiceCartRentalBody struct {
	PageRequest
	PartnerUid   string `form:"partner_uid"`
	CourseUid    string `form:"course_uid"`
	BookingDate  string `form:"booking_date" binding:"required"`
	ServiceId    int64  `form:"service_id"`
	RentalStatus string `form:"rental_status"`
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
	PartnerUid     string  `json:"partner_uid"`
	CourseUid      string  `json:"course_uid"`
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
