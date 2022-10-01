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
	GolfBag       string              `json:"golf_bag" binding:"required"`
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
}

type UpdateResItemBody struct {
	ItemId int64 `json:"item_id" binding:"required"`
}

type GetFoodProcessBody struct {
	ServiceId int64  `json:"service_id" binding:"required"`
	Type      string `json:"type"`
	Name      string `json:"name"`
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
