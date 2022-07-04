package request

type CreateFoodBeverageBody struct {
	PartnerUid    string  `json:"partner_uid"`
	CourseUid     string  `json:"course_uid"`
	GroupId       string  `json:"group_id"`
	GroupName     string  `json:"group_name"`
	FBCode        string  `json:"fb_code"`
	EnglishName   string  `json:"english_name"`
	VieName       string  `json:"vietnamese_name"`
	Unit          string  `json:"unit"`
	Price         float64 `json:"price"`
	NetCost       float64 `json:"net_cost"`
	CostPrice     float64 `json:"cost_price"`
	Barcode       string  `json:"barcode"`
	AccountCode   string  `json:"account_code"`
	BarBeerPrice  float64 `json:"bar_beer_price"`
	Note          string  `json:"note"`
	ForKiosk      bool    `json:"for_kiosk"`
	OpenFB        float64 `json:"open_fb"`
	AloneKiosk    string  `json:"alone_kiosk"`
	InMenuSet     bool    `json:"in_menu_set"`    // Món trong combo
	IsInventory   bool    `json:"is_inventory"`   // Có trong kho
	InternalPrice float64 `json:"internal_price"` // Giá nội bộ là giá dành cho nhân viên ăn uống và sử dụng
}

type GetListFoodBeverageForm struct {
	PageRequest
	PartnerUid         *string `form:"partner_uid" json:"partner_uid"`
	CourseUid          *string `form:"course_uid" json:"course_uid"`
	EnglishName        *string `form:"english_name" json:"english_name"`
	VieName            *string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupId            *string `form:"group_id" json:"group_id"`
	FoodBeverageStatus *string `form:"rental_status" json:"rental_status"`
}

type UpdateFoodBeverageBody struct {
	EnglishName        *string `json:"english_name"`
	VieName            *string `json:"vietnamese_name"`
	FoodBeverageStatus *string `json:"rental_status"`
	ByHoles            *bool   `json:"by_holes"`
	ForPos             *bool   `json:"for_pos"`
	OnlyForRen         *bool   `json:"only_for_ren"`
	Price              *int64  `json:"price"`
}
