package request

type CreateFoodBeverageBody struct {
	PartnerUid    string  `json:"partner_uid" binding:"required"`
	CourseUid     string  `json:"course_uid" binding:"required"`
	GroupCode     string  `json:"group_code" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	GroupName     string  `json:"group_name"`
	FBCode        string  `json:"fb_code"`
	AccountCode   string  `json:"account_code"`
	EnglishName   string  `json:"english_name"`
	VieName       string  `json:"vietnamese_name"`
	Unit          string  `json:"unit"`
	Price         float64 `json:"price"`
	NetCost       float64 `json:"net_cost"`
	CostPrice     float64 `json:"cost_price"`
	Barcode       string  `json:"barcode"`
	BarBeerPrice  float64 `json:"bar_beer_price"`
	Note          string  `json:"note"`
	ForKiosk      bool    `json:"for_kiosk"`
	OpenFB        bool    `json:"open_fb"`
	AloneKiosk    string  `json:"alone_kiosk"`
	InMenuSet     bool    `json:"in_menu_set"`
	IsInventory   bool    `json:"is_inventory"`
	InternalPrice float64 `json:"internal_price"`
	IsKitchen     bool    `json:"is_kitchen"`
	Status        string  `json:"status"`
	HotKitchen    *bool   `json:"hot_kitchen"`
	ColdKitchen   *bool   `json:"cold_kitchen"`
}

type GetListFoodBeverageForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" json:"partner_uid"`
	CourseUid   string `form:"course_uid" json:"course_uid"`
	EnglishName string `form:"english_name" json:"english_name"`
	VieName     string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupCode   string `form:"group_code" json:"group_code"`
	FBCode      string `form:"fb_code" json:"fb_code"`
	Status      string `form:"status" json:"status"`
	FBCodeList  string `form:"fb_code_list" json:"fb_code_list"`
	Type        string `form:"type"`
	CodeOrName  string `form:"code_or_name"`
}

type UpdateFoodBeverageBody struct {
	EnglishName   string  `json:"english_name"`
	VieName       string  `json:"vietnamese_name"`
	Unit          string  `json:"unit"`
	Price         float64 `json:"price"`
	NetCost       float64 `json:"net_cost"`
	CostPrice     float64 `json:"cost_price"`
	BarBeerPrice  float64 `json:"bar_beer_price"`
	InternalPrice float64 `json:"internal_price"`
	Barcode       string  `json:"barcode"`
	AccountCode   string  `json:"account_code"`
	Note          string  `json:"note"`
	Status        string  `json:"status"`
	AloneKiosk    string  `json:"alone_kiosk"`
	ForKiosk      *bool   `json:"for_kiosk"`
	OpenFB        *bool   `json:"open_fb"`
	InMenuSet     *bool   `json:"in_menu_set"`
	IsInventory   *bool   `json:"is_inventory"`
	IsKitchen     *bool   `json:"is_kitchen"`
	UserUpdate    string  `json:"user_update"`
	HotKitchen    *bool   `json:"hot_kitchen"`
	ColdKitchen   *bool   `json:"cold_kitchen"`
	Type          string  `json:"type"`
	GroupCode     string  `json:"group_code"`
	GroupName     string  `json:"group_name"`
}
