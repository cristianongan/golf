package request

type KioskInventoryInputItemBody struct {
	PartnerUid    string  `json:"partner_uid" binding:"required"`
	CourseUid     string  `json:"course_uid" binding:"required"`
	Code          string  `json:"code" binding:"required"`
	ItemCode      string  `json:"item_code" binding:"required"`
	ItemName      string  `json:"item_name" binding:"required"`
	GoodsCode     string  `json:"goods_code" binding:"required"`
	Quantity      int64   `json:"quantity" binding:"required"`
	Source        string  `json:"source"`
	ReviewUserUid string  `json:"review_user_uid"`
	Note          string  `json:"note"`
	KioskCode     string  `json:"kiosk_code" binding:"required"`
	KioskName     string  `json:"kiosk_name" binding:"required"`
	Price         float64 `json:"price" binding:"required"`
}

type KioskInventoryOutputItemBody struct {
	PartnerUid    string  `json:"partner_uid" binding:"required"`
	CourseUid     string  `json:"course_uid" binding:"required"`
	Code          string  `json:"code" binding:"required"`
	ItemCode      string  `json:"item_code" binding:"required"`
	ItemName      string  `json:"item_name" binding:"required"`
	GoodsCode     string  `json:"Goods_code" binding:"required"`
	Quantity      int64   `json:"quantity" binding:"required"`
	ReviewUserUid string  `json:"review_user_uid"`
	Note          string  `json:"note"`
	KioskCode     string  `json:"kiosk_code" binding:"required"`
	KioskName     string  `json:"kiosk_name" binding:"required"`
	Source        string  `json:"source" binding:"required"`
	Price         float64 `json:"price" binding:"required"`
}

type CreateKioskInventoryBillBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Code       string `json:"code" binding:"required"`
	KioskCode  string `json:"kiosk_code" binding:"required"`
	KioskName  string `json:"kiosk_name" binding:"required"`
	Source     string `json:"source"`
}

type GetInOutItems struct {
	PageRequest
	KioskCode  string `form:"kiosk_code" binding:"required"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
}

type GetBill struct {
	PageRequest
	BillStatus string `form:"bill_status"`
	KioskCode  string `form:"kiosk_code" binding:"required"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"`
}

type KioskInventoryInsertBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Code       string `json:"code" binding:"required"` // Mã đơn nhập
	KioskCode  string `json:"kiosk_code" binding:"required"`
}
