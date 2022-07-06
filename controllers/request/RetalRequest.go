package request

type CreateRentalBody struct {
	RentalId    string `json:"rental_id"`
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	EnglishName string `json:"english_name"`
	VieName     string `json:"vietnamese_name"`
	RenPos      string `json:"ren_pos"`
	SystemCode  string `json:"system_code"`
	GroupCode   string `json:"group_code"`
	Unit        string `json:"unit"`
	Price       int64  `json:"price"`
	ByHoles     bool   `json:"by_holes"`
	ForPos      bool   `json:"for_pos"`
	OnlyForRen  bool   `json:"only_for_ren"`
	InputUser   string `json:"input_user"`
	Status      string `json:"status"`
}

type GetListRentalForm struct {
	PageRequest
	PartnerUid  *string `form:"partner_uid" json:"partner_uid"`
	CourseUid   *string `form:"course_uid" json:"course_uid"`
	EnglishName *string `form:"english_name" json:"english_name"`
	VieName     *string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupCode   *string `form:"group_code" json:"group_code"`
	Status      *string `json:"status"`
}

type UpdateRentalBody struct {
	EnglishName *string `json:"english_name"`
	VieName     *string `json:"vietnamese_name"`
	RenPos      *string `json:"ren_pos"`
	SystemCode  *string `json:"system_code"`
	GroupCode   *string `json:"group_code"`
	Unit        *string `json:"unit"`
	Price       *int64  `json:"price"`
	ByHoles     *bool   `json:"by_holes"`
	ForPos      *bool   `json:"for_pos"`
	OnlyForRen  *bool   `json:"only_for_ren"`
	InputUser   *string `json:"input_user"`
	Status      *string `json:"status"`
}
