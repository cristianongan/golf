package request

type CreateRentalBody struct {
	RentalId    string  `json:"rental_id" binding:"required"`
	PartnerUid  string  `json:"partner_uid" binding:"required"`
	CourseUid   string  `json:"course_uid" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	GroupCode   string  `json:"group_code"`
	EnglishName string  `json:"english_name"`
	VieName     string  `json:"vietnamese_name"`
	RenPos      string  `json:"ren_pos"`
	Unit        string  `json:"unit"`
	Price       float64 `json:"price"`
	ByHoles     bool    `json:"by_holes"`
	ForPos      bool    `json:"for_pos"`
	OnlyForRen  bool    `json:"only_for_ren"`
	InputUser   string  `json:"input_user"`
	Status      string  `json:"status"`
	IsDriving   *bool   `json:"is_driving"`
	Rate        string  `json:"rate"`
	AccountCode string  `json:"account_code"`
	TaxCode     string  `json:"tax_code"`
}

type GetListRentalForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" json:"partner_uid"`
	CourseUid   string `form:"course_uid" json:"course_uid"`
	EnglishName string `form:"english_name" json:"english_name"`
	VieName     string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupCode   string `form:"group_code" json:"group_code"`
	Status      string `json:"status"`
	Type        string `form:"type"`
	CodeOrName  string `form:"code_or_name"`
	IsDriving   *bool  `form:"is_driving"`
	GuestStyle  string `form:"guest_style"`
}

type UpdateRentalBody struct {
	EnglishName string   `json:"english_name"`
	VieName     string   `json:"vietnamese_name"`
	RenPos      string   `json:"ren_pos"`
	SystemCode  string   `json:"system_code"`
	Unit        string   `json:"unit"`
	Price       *float64 `json:"price"`
	ByHoles     *bool    `json:"by_holes"`
	ForPos      *bool    `json:"for_pos"`
	OnlyForRen  *bool    `json:"only_for_ren"`
	InputUser   string   `json:"input_user"`
	Status      string   `json:"status"`
	Type        string   `json:"type"`
	GroupCode   string   `json:"group_code"`
	GroupName   string   `json:"group_name"`
	IsDriving   *bool    `json:"is_driving"`
	Rate        string   `json:"rate"`
	TaxCode     string   `json:"tax_code"`
}
