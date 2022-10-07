package request

type CreateSinglePaymentBody struct {
	BookingUid  string `json:"booking_uid" binding:"required"` // Booking uid
	BillCode    string `json:"bill_code" binding:"required"`
	DateStr     string `json:"date_str" binding:"required"`  // timestamp hiện tại -> string
	PaymentType string `json:"payment_type"`                 // CASH, VISA
	Amount      int64  `json:"amount" binding:"required"`    // Số tiền thanh toán
	CheckSum    string `json:"check_sum" binding:"required"` // Checksum
	Note        string `json:"note"`                         // Note
}

type GetListSinglePaymentBody struct {
	PageRequest
	PartnerUid    string `json:"partner_uid" binding:"required"`
	CourseUid     string `json:"course_uid"`
	Bag           string `json:"bag"`
	PlayerName    string `json:"player_name"`
	PaymentStatus string `json:"payment_status"`
	PaymentDate   string `json:"payment_date"`
	CheckSum      string `json:"check_sum" binding:"required"` // Checksum
}

type UpdateSinglePaymentItemBody struct {
	BookingUid     string `json:"booking_uid" binding:"required"` // Booking uid
	PaymentItemUid string `json:"payment_item_uid" binding:"required"`
	DateStr        string `json:"date_str" binding:"required"`  // timestamp hiện tại -> string
	CheckSum       string `json:"check_sum" binding:"required"` // Checksum
	Note           string `json:"note"`                         // Note
}

type GetListSinglePaymentDetailBody struct {
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid"`
	BillCode    string `json:"bill_code" binding:"required"`
	Bag         string `json:"bag" binding:"required"`
	PaymentDate string `json:"payment_date"`
	CheckSum    string `json:"check_sum" binding:"required"` // Checksum
}

type DeleteSinglePaymentDetailBody struct {
	SinglePaymentItemUid string `json:"single_payment_item_uid" binding:"required"`
	BillCode             string `json:"bill_code" binding:"required"`
	Bag                  string `json:"bag" binding:"required"`
	PaymentDate          string `json:"payment_date"`
	CheckSum             string `json:"check_sum" binding:"required"` // Checksum
}
