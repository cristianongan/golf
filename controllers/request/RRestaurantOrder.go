package request

type CreateRestaurantOrderBody struct {
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	GolfBag     string `json:"golf_bag"`
	ServiceId   int64  `json:"service_id"`
	Type        string `json:"type"`
	TypeCode    string `json:"type_code"`
	NumberGuest int    `json:"number_guest"`
}

type AddItemOrderBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	BillId     int64  `json:"bill_id"`
	ItemCode   string `json:"item_code"`
	Type       string `json:"type"`
	Quantity   int    `json:"quantity"`
}

type UpdateItemOrderBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	ItemId     int64  `json:"item_id"`
	Quantity   int    `json:"quantity"`
	Note       string `json:"note"`
}

type CreateBillOrderBody struct {
	BillId int64 `json:"bill_id"`
}

type GetItemResOrderBody struct {
	PageRequest
	BillId int64 `form:"bill_id"`
}

type GetListBillBody struct {
	PageRequest
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	BookingDate string `form:"booking_date" binding:"required"`
	ServiceId   int64  `form:"service_id"`
	BillStatus  string `form:"bill_status"`
}

type UpdateResItemBody struct {
	ItemId int64 `json:"item_id"`
}

type GetFoodProcessBody struct {
	ServiceId int64  `json:"service_id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
}

type GetDetailFoodProcessBody struct {
	ServiceId int64  `json:"service_id"`
	ItemCode  string `json:"item_code"`
}

type FinishAllResItemBody struct {
	ServiceId int64  `json:"service_id"`
	BillId    int64  `json:"bill_id"`
	ItemCode  string `json:"item_code"`
}
