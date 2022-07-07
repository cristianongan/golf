package request

type CreateDepositBody struct {
	PartnerUid     string  `json:"partner_uid"`
	CourseUid      string  `json:"course_uid"`
	InputDate      string  `json:"input_date"`
	CustomerUid    string  `json:"customer_uid"`
	PaymentType    string  `json:"payment_type"`
	TmCurrency     string  `json:"tm_currency"`
	TmRate         float64 `json:"tm_rate"`
	TmOriginAmount int64   `json:"tm_origin_amount"`
	TcCurrency     string  `json:"tc_currency"`
	TcRate         float64 `json:"tc_rate"`
	TcOriginAmount int64   `json:"tc_origin_amount"`
	Note           string  `json:"note"`
}

type GetDepositList struct {
	PageRequest
	CustomerIdentity string `form:"customer_identity"`
	CustomerPhone    string `form:"customer_phone"`
	CustomerStyle    string `form:"customer_style"`
	InputDate        string `form:"input_date"`
}

type UpdateDepositBody struct {
	PaymentType    string  `json:"payment_type"`
	TmCurrency     string  `json:"tm_currency"`
	TmRate         float64 `json:"tm_rate"`
	TmOriginAmount int64   `json:"tm_origin_amount"`
	TcCurrency     string  `json:"tc_currency"`
	TcRate         float64 `json:"tc_rate"`
	TcOriginAmount int64   `json:"tc_origin_amount"`
	Note           string  `json:"note"`
}
