package request

type CreateFoodBeverageBody struct {
	PartnerUid    string  `json:"partner_uid"`
	CourseUid     string  `json:"course_uid"`
	GroupCode     string  `json:"group_code"`
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
	OpenFB        bool    `json:"open_fb"`
	AloneKiosk    string  `json:"alone_kiosk"`
	InMenuSet     bool    `json:"in_menu_set"`
	IsInventory   bool    `json:"is_inventory"`
	InternalPrice float64 `json:"internal_price"`
	IsKitchen     bool    `json:"is_kitchen"`
	Status        string  `json:"status"`
}

type GetListFoodBeverageForm struct {
	PageRequest
	PartnerUid  *string `form:"partner_uid" json:"partner_uid"`
	CourseUid   *string `form:"course_uid" json:"course_uid"`
	EnglishName *string `form:"english_name" json:"english_name"`
	VieName     *string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupCode   *string `form:"group_code" json:"group_code"`
	Status      *string `form:"status" json:"status"`
}

type UpdateFoodBeverageBody struct {
	GroupCode     *string  `json:"group_code"`
	EnglishName   *string  `json:"english_name"`
	VieName       *string  `json:"vietnamese_name"`
	Unit          *string  `json:"unit"`
	Price         *float64 `json:"price"`
	NetCost       *float64 `json:"net_cost"`
	CostPrice     *float64 `json:"cost_price"`
	BarBeerPrice  *float64 `json:"bar_beer_price"`
	InternalPrice *float64 `json:"internal_price"`
	Barcode       *string  `json:"barcode"`
	AccountCode   *string  `json:"account_code"`
	Note          *string  `json:"note"`
	Status        *string  `json:"status"`
	AloneKiosk    *string  `json:"alone_kiosk"`
	ForKiosk      *bool    `json:"for_kiosk"`
	OpenFB        *bool    `json:"open_fb"`
	InMenuSet     *bool    `json:"in_menu_set"`
	IsInventory   *bool    `json:"is_inventory"`
	IsKitchen     *bool    `json:"is_kitchen"`
}
