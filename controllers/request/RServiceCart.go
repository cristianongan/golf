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
	PartnerUid   string `json:"partner_uid"`
	CourseUid    string `json:"course_uid"`
	GolfBag      string `json:"golf_bag" binding:"required"`
	ItemCode     string `json:"item_code"`
	Quantity     int64  `json:"quantity"`
	ServiceId    int64  `json:"service_id"`
	BillId       int64  `json:"bill_id"`
	Name         string `json:"name"`
	Price        int64  `json:"price"`
	Hole         int    `json:"hole"`
	CaddieCode   string `json:"caddie_code"`
	ServiceType  string `json:"service_type"`
	LocationType string `json:"location_type"`
}

type AddDiscountServiceItemBody struct {
	ItemId         int64  `json:"item_id"`
	DiscountType   string `json:"discount_type"`
	DiscountPrice  int64  `json:"discount_price"`
	DiscountReason string `json:"discount_reason"`
}

type AddDiscountBillBody struct {
	BillId         int64  `json:"bill_id"`
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
	GolfBag     string `form:"golf_bag"`
	UserName    string `form:"user_name"`
}

type GetServiceCartRentalBody struct {
	PageRequest
	PartnerUid   string `form:"partner_uid"`
	CourseUid    string `form:"course_uid"`
	BookingDate  string `form:"booking_date" binding:"required"`
	ServiceId    int64  `form:"service_id"`
	RentalStatus string `form:"rental_status"`
	GolfBag      string `form:"golf_bag"`
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
	PartnerUid     string  `json:"partner_uid" binding:"required"`
	CourseUid      string  `json:"course_uid" binding:"required"`
	ServiceCartId  int64   `json:"service_cart_id" binding:"required"`
	GolfBag        string  `json:"golf_bag" binding:"required"`
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

type SaveBillPOSInAppBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	GolfBag    string `json:"golf_bag" binding:"required"`
	ServiceId  int64  `json:"service_id"`
	BillId     int64  `json:"bill_id"`
	Note       string `json:"note"`
	// Infor nhà hàng
	Type        string `json:"type"`
	TypeCode    string `json:"type_code"`
	NumberGuest int    `json:"number_guest"`
	Floor       int    `json:"floor"`
	// Infor item
	Items []Item `json:"items"`
}

type Item struct {
	Action        string `json:"action" binding:"required"`
	ItemCode      string `json:"item_code" binding:"required"`
	ItemId        int64  `json:"item_id"`
	Type          string `json:"type"`
	Quantity      int    `json:"quantity"`
	UnitPrice     int64  `json:"unit_price"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	Note          string `json:"note"`
}
