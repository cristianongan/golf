package request

type CreateProshopBody struct {
	ProshopId     string  `json:"proshop_id" binding:"required"`
	PartnerUid    string  `json:"partner_uid" binding:"required"`
	CourseUid     string  `json:"course_uid" binding:"required"`
	GroupCode     string  `json:"group_code" binding:"required"`
	AccountCode   string  `json:"account_code" binding:"required"`
	Brand         string  `json:"brand"`
	EnglishName   string  `json:"english_name"`
	VieName       string  `json:"vietnamese_name"`
	Unit          string  `json:"unit"`
	Price         float64 `json:"price"`
	NetCost       float64 `json:"net_cost"`
	CostPrice     float64 `json:"cost_price"`
	Barcode       string  `json:"barcode"`
	Note          string  `json:"note" `
	ForKiosk      bool    `json:"for_kiosk"`
	ProPrice      float64 `json:"pro_price"`
	IsInventory   bool    `json:"is_inventory"`
	Type          string  `json:"type"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	UserUpdate    string  `json:"user_update"`
	IsDeposit     bool    `json:"is_deposit"`
	PeopleDeposit string  `json:"people_deposit"`
}

type GetListProshopForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" json:"partner_uid"`
	CourseUid   string `form:"course_uid" json:"course_uid"`
	EnglishName string `form:"english_name" json:"english_name"`
	VieName     string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupCode   string `form:"group_code" json:"group_code"`
	GroupName   string `form:"group_name" json:"group_name"`
}
type UpdateProshopBody struct {
	GroupCode     *string  `json:"group_code"`
	Brand         *string  `json:"brand"`
	EnglishName   *string  `json:"english_name"`
	VieName       *string  `json:"vietnamese_name"`
	Unit          *string  `json:"unit"`
	Price         *float64 `json:"price"`
	NetCost       *float64 `json:"net_cost"`
	CostPrice     *float64 `json:"cost_price"`
	Barcode       *string  `json:"barcode"`
	AccountCode   *string  `json:"account_code"`
	Note          *string  `json:"note" `
	ForKiosk      *bool    `json:"for_kiosk"`
	ProPrice      *float64 `json:"pro_price"`
	IsInventory   *bool    `json:"is_inventory"`
	Type          *string  `json:"type"`
	Code          *string  `json:"code"`
	Name          *string  `json:"name"`
	UserUpdate    *string  `json:"user_update"`
	IsDeposit     *bool    `json:"is_deposit"`
	PeopleDeposit *string  `json:"people_deposit"`
}
