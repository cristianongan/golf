package request

type CreateCaddieConfigFeeBody struct {
	PartnerUid string `json:"partner_uid" validate:"required"`
	CourseUid  string `json:"course_uid" validate:"required"`
	Type       string `json:"type" validate:"required"`
	FeeDetail  string `json:"fee_detail" validate:"required"`
	ValidDate  string `json:"valid_date" validate:"required"`
	ExpDate    string `json:"exp_date" validate:"required"`
}

type GetCaddieConfigFeeList struct {
	PageRequest
	Type      *string `form:"type"`
	ValidDate *string `form:"valid_date"`
	ExpDate   *string `form:"exp_date"`
}

type UpdateCaddieConfigFeeBody struct {
	PartnerUid *string `json:"partner_uid" validate:"required"`
	CourseUid  *string `json:"course_uid" validate:"required"`
	Type       *string `json:"type" validate:"required"`
	FeeDetail  *string `json:"fee_detail" validate:"required"`
	ValidDate  *string `json:"valid_date" validate:"required"`
	ExpDate    *string `json:"exp_date" validate:"required"`
}
