package request

type CreateProshopBody struct {
	PartnerUid  string  `json:"partner_uid"` // Hang Golf
	CourseUid   string  `json:"course_uid"`  // San Golf
	GroupId     string  `json:"group_id"`
	Brand       string  `json:"brand"`
	GroupName   string  `json:"group_name"`
	ProCode     string  `json:"pro_code"`
	EnglishName string  `json:"english_name"`    // Tên Tiếng Anh
	VieName     string  `json:"vietnamese_name"` // Tên Tiếng Anh
	Unit        string  `json:"unit"`
	Price       float64 `json:"price"`
	NetCost     float64 `json:"net_cost"` // Net cost tự tính từ Cost Price ko bao gồm 10% VAT
	CostPrice   float64 `json:"cost_price"`
	Barcode     string  `json:"barcode"`
	AccountCode string  `json:"account_code"` // Mã liên kết với Account kế toán
	Note        string  `json:"note" `
	ForKiosk    bool    `json:"for_kiosk"`
	ProPrice    float64 `json:"pro_price"`
	IsInventory bool    `json:"is_inventory"` // Có trong kho
	Type        string  `json:"type"`         // Loại rental, kiosk, proshop,...
	Code        string  `json:"code"`
	Name        string  `json:"name"`        // Tên
	UserUpdate  string  `json:"user_update"` // Tên
}

type GetListProshopForm struct {
	PageRequest
	PartnerUid         *string `form:"partner_uid" json:"partner_uid"`
	CourseUid          *string `form:"course_uid" json:"course_uid"`
	EnglishName        *string `form:"english_name" json:"english_name"`
	VieName            *string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupId            *string `form:"group_id" json:"group_code"`
	FoodBeverageStatus *string `form:"rental_status" json:"rental_status"`
}

type UpdateProshopBody struct {
	EnglishName        *string `json:"english_name"`
	VieName            *string `json:"vietnamese_name"`
	FoodBeverageStatus *string `json:"rental_status"`
	ByHoles            *bool   `json:"by_holes"`
	ForPos             *bool   `json:"for_pos"`
	OnlyForRen         *bool   `json:"only_for_ren"`
	Price              *int64  `json:"price"`
}
