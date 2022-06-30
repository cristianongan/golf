package request

type CreateRentalBody struct {
	PartnerUid   string `json:"partner_uid"`                              // Hang Golf
	CourseUid    string `json:"course_uid"`                               // San Golf
	EnglishName  string `json:"english_name"`                             // Tên Tiếng Anh
	VieName      string `json:"vietnamese_name" gorm:"type:varchar(256)"` // Tên Tiếng Anh
	RenPos       string `json:"ren_pos" gorm:"type:varchar(100)"`
	Code         string `json:"code" gorm:"type:varchar(100)"`
	GroupId      int64  `json:"group_id" gorm:"index"`
	GroupCode    string `json:"group_code" gorm:"type:varchar(100);index"`
	GroupName    string `json:"group_name" gorm:"type:varchar(256)"`
	Unit         string `json:"unit" gorm:"type:varchar(100)"`
	Price        int64  `json:"price"`
	ByHoles      bool   `json:"by_holes"`
	ForPos       bool   `json:"for_pos"`
	OnlyForRen   bool   `json:"only_for_ren"`
	RentalStatus string `json:"rental_status" gorm:"type:varchar(100)"`
	InputUser    string `json:"input_user" gorm:"type:varchar(100)"`
}

type GetListRentalForm struct {
	PageRequest
	PartnerUid   *string `form:"partner_uid" json:"partner_uid"`
	CourseUid    *string `form:"course_uid" json:"course_uid"`
	EnglishName  *string `form:"english_name" json:"english_name"`
	VieName      *string `form:"vietnamese_name" json:"vietnamese_name"`
	GroupCode    *string `form:"group_code" json:"group_code"`
	RentalStatus *string `form:"rental_status" json:"rental_status"`
}

type UpdateRentalBody struct {
	EnglishName  *string `json:"english_name"`
	VieName      *string `json:"vietnamese_name"`
	RentalStatus *string `json:"rental_status"`
	ByHoles      *bool   `json:"by_holes"`
	ForPos       *bool   `json:"for_pos"`
	OnlyForRen   *bool   `json:"only_for_ren"`
	Price        *int64  `json:"price"`
}
