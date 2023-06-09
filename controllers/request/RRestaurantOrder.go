package request

import "start/utils"

type CreateRestaurantOrderBody struct {
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	GolfBag     string `json:"golf_bag" binding:"required"`
	ServiceId   int64  `json:"service_id" binding:"required"`
	Type        string `json:"type" binding:"required"`
	TypeCode    string `json:"type_code" binding:"required"`
	NumberGuest int    `json:"number_guest"`
	Floor       int    `json:"floor"`
}

type CreateBookingRestaurantBody struct {
	PartnerUid    string              `json:"partner_uid"`
	CourseUid     string              `json:"course_uid"`
	GolfBag       string              `json:"golf_bag"`
	ServiceId     int64               `json:"service_id" binding:"required"`
	FromServiceId int64               `json:"from_service_id"`
	PlayerName    string              `json:"player_name" binding:"required"`
	Phone         string              `json:"phone" binding:"required"`
	OrderTime     int64               `json:"order_time"`
	Table         string              `json:"table"`
	NumberGuest   int                 `json:"number_guest"`
	Floor         int                 `json:"floor"`
	ListOrderItem utils.ListOrderItem `json:"list_order_item"`
	Note          string              `json:"note"`
}

type UpdateBookingRestaurantBody struct {
	PartnerUid    string              `json:"partner_uid"`
	CourseUid     string              `json:"course_uid"`
	GolfBag       string              `json:"golf_bag"`
	PlayerName    string              `json:"player_name" binding:"required"`
	Phone         string              `json:"phone" binding:"required"`
	OrderTime     int64               `json:"order_time"`
	Table         string              `json:"table"`
	NumberGuest   int                 `json:"number_guest"`
	Floor         int                 `json:"floor"`
	ListOrderItem utils.ListOrderItem `json:"list_order_item"`
	Note          string              `json:"note"`
}

type ConfrimBookingRestaurantBody struct {
	GolfBag string `json:"golf_bag"`
}

type AddItemOrderBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	BillId     int64  `json:"bill_id" binding:"required"`
	ItemCode   string `json:"item_code" binding:"required"`
	Type       string `json:"type"`
	Quantity   int    `json:"quantity"`
}

type UpdateItemOrderBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	ItemId     int64  `json:"item_id" binding:"required"`
	Quantity   int    `json:"quantity"`
	Note       string `json:"note"`
}

type CreateBillOrderBody struct {
	BillId int64 `json:"bill_id" binding:"required"`
}

type GetItemResOrderBody struct {
	PageRequest
	BillId int64 `form:"bill_id" binding:"required"`
}

type GetListBillBody struct {
	PageRequest
	PartnerUid   string `json:"partner_uid"`
	CourseUid    string `json:"course_uid"`
	BookingDate  string `form:"booking_date" binding:"required"`
	ServiceId    int64  `form:"service_id" binding:"required"`
	BillStatus   string `form:"bill_status"`
	BillCode     string `form:"bill_code"`
	CustomerName string `form:"customer_name"`
	GolfBag      string `form:"golf_bag"`
	Table        string `form:"table"`
	Type         string `form:"type"`
	Floor        int    `form:"floor"`
	FromService  int64  `form:"from_service"`
}

type UpdateResItemBody struct {
	ItemCode string `json:"item_code" binding:"required"`
	BillId   int64  `json:"bill_id" binding:"required"`
}

type GetFoodProcessBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	ServiceId  int64  `json:"service_id" binding:"required"`
	OrderDate  string `json:"order_date" binding:"required"`
	Type       string `json:"type"`
	Status     string `json:"status"`
}

type GetDetailFoodProcessBody struct {
	ServiceId int64  `json:"service_id" binding:"required"`
	ItemCode  string `json:"item_code" binding:"required"`
}

type FinishAllResItemBody struct {
	ServiceId int64  `json:"service_id" binding:"required"`
	BillId    int64  `json:"bill_id"`
	ItemCode  string `json:"item_code"`
}

type FinishRestaurantOrderBody struct {
	BillId int64 `json:"bill_id" binding:"required"`
}

type TransferItemBody struct {
	PartnerUid     string  `json:"partner_uid" binding:"required"`
	CourseUid      string  `json:"course_uid" binding:"required"`
	ServiceCartId  int64   `json:"service_cart_id" binding:"required"`
	GolfBag        string  `json:"golf_bag" binding:"required"`
	CartItemIdList []int64 `json:"cart_item_id_list"`
}

type ActionKitchenBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	ItemCode   string `json:"item_code" binding:"required"`
	OrderDate  string `json:"order_date" binding:"required"`
	ServiceId  int64  `json:"service_id" binding:"required"`
	Type       string `json:"type"`
	Action     string `json:"action"`
	Group      string `json:"group"`
	//
	BillId         int64 `json:"bill_id"`
	QuantityReturn int   `json:"quantity_return"`
}
