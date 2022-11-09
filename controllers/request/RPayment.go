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
	BookingUid           string `json:"booking_uid" binding:"required"` // Booking uid
	SinglePaymentItemUid string `json:"single_payment_item_uid" binding:"required"`
	DateStr              string `json:"date_str" binding:"required"`  // timestamp hiện tại -> string
	CheckSum             string `json:"check_sum" binding:"required"` // Checksum
	Note                 string `json:"note"`                         // Note
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
	CheckSum             string `json:"check_sum" binding:"required"` // Checksum
}

type CreateAgencyPaymentItemBody struct {
	AgencyPaymentUid string `json:"agency_payment_uid" binding:"required"` // Booking uid
	BookingCode      string `json:"booking_code" binding:"required"`       // Booking uid
	DateStr          string `json:"date_str" binding:"required"`           // timestamp hiện tại -> string
	PaymentType      string `json:"payment_type"`                          // CASH, VISA, DEBT
	Amount           int64  `json:"amount" binding:"required"`             // Số tiền thanh toán
	CheckSum         string `json:"check_sum" binding:"required"`          // Checksum
	Note             string `json:"note"`                                  // Note
}

type GetListAgencyPaymentBody struct {
	PageRequest
	PartnerUid    string `json:"partner_uid" binding:"required"`
	CourseUid     string `json:"course_uid"`
	Bag           string `json:"bag"`
	PlayerName    string `json:"player_name"`
	AgencyName    string `json:"agency_name"`
	PaymentStatus string `json:"payment_status"`
	PaymentDate   string `json:"payment_date"`
	CheckSum      string `json:"check_sum" binding:"required"` // Checksum
}

type DeleteAgencyPaymentDetailBody struct {
	AgencyPaymentItemUid string `json:"agency_payment_item_uid" binding:"required"`
	BookingCode          string `json:"booking_code" binding:"required"`
	CheckSum             string `json:"check_sum" binding:"required"` // Checksum
}

type GetListAgencyPaymentItemBody struct {
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid"`
	BookingCode string `json:"booking_code" binding:"required"`
	PaymentUid  string `json:"payment_uid" binding:"required"`
	CheckSum    string `json:"check_sum" binding:"required"` // Checksum
}

type GetAgencyPayForBagDetailBody struct {
	BookingCode string `json:"booking_code" binding:"required"`
	BookingUid  string `json:"booking_uid" binding:"required"`
	CheckSum    string `json:"check_sum" binding:"required"` // Checksum
}
