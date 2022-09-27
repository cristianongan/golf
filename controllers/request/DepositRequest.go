package request

type CreateDepositBody struct {
	//PartnerUid     string  `json:"partner_uid"`
	//CourseUid      string  `json:"course_uid"`
	InputDate        string  `json:"input_date" validate:"required"`
	CustomerUid      string  `json:"customer_uid"`
	PaymentType      string  `json:"payment_type" validate:"required"`
	TmCurrency       string  `json:"tm_currency" validate:"required_if=PaymentType TM,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TmRate           float64 `json:"tm_rate" validate:"required_if=PaymentType TM,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TmOriginAmount   int64   `json:"tm_origin_amount" validate:"required_if=PaymentType TM,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TcCurrency       string  `json:"tc_currency" validate:"required_if=PaymentType CC,required_if=PaymentType CK,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TcRate           float64 `json:"tc_rate" validate:"required_if=PaymentType CC,required_if=PaymentType CK,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TcOriginAmount   int64   `json:"tc_origin_amount" validate:"required_if=PaymentType CC,required_if=PaymentType CK,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	Note             string  `json:"note"`
	CustomerName     string  `json:"customer_name"`
	CustomerPhone    string  `json:"customer_phone"`
	CustomerIdentity string  `json:"customer_identity"`
}

type GetDepositList struct {
	PageRequest
	CustomerIdentity string `form:"customer_identity"`
	CustomerPhone    string `form:"customer_phone"`
	CustomerStyle    string `form:"customer_style"`
	InputDate        string `form:"input_date"`
}

type UpdateDepositBody struct {
	CustomerUid    string  `json:"customer_uid" validate:"required"`
	PaymentType    string  `json:"payment_type" validate:"required"`
	TmCurrency     string  `json:"tm_currency" validate:"required_if=PaymentType TM,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TmRate         float64 `json:"tm_rate" validate:"required_if=PaymentType TM,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TmOriginAmount int64   `json:"tm_origin_amount" validate:"required_if=PaymentType TM,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TcCurrency     string  `json:"tc_currency" validate:"required_if=PaymentType CC,required_if=PaymentType CK,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TcRate         float64 `json:"tc_rate" validate:"required_if=PaymentType CC,required_if=PaymentType CK,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	TcOriginAmount int64   `json:"tc_origin_amount" validate:"required_if=PaymentType CC,required_if=PaymentType CK,required_if=PaymentType TMCC,required_if=PaymentType TMCK"`
	Note           string  `json:"note"`
}
