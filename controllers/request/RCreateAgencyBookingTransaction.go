package request

import models_agency_booking "start/models/agency-booking"

type AgencyBookingTransactionDTO struct {
	TransactionId        string `json:"transaction_id"`         // mã giao dịch
	CourseUid            string `json:"course_id"`              // mã sân
	PartnerUid           string `json:"partner_id"`             //
	AgencyId             int64  `json:"agency_id"`              // Mã đại lí
	PaymentStatus        string `json:"payment_status"`         // trạng thái thanh toán
	TransactionStatus    string `json:"transaction_status"`     // trạng thái đơn ~ trạng thái giao dịch
	BookingRequestStatus string `json:"booking_request_status"` // trạng thái yêu cầu booking với nhà cung cấp
	CustomerPhoneNumber  string `json:"customer_phone_number"`  // điện thoại người đặt
	CustomerEmail        string `json:"customer_email"`         // email người đặt
	CustomerName         string `json:"customer_name"`          // Tên người đặt
	BookingAmount        int64  `json:"booking_amount"`         // giá book
	ServiceAmount        int64  `json:"service_amount"`         // giá dịch vụ
	PaymentDueDate       int64  `json:"payment_due_date"`       // ngày hết hạn thanh toán
	PlayerNote           string `json:"player_note"`            // ghi ch ú người chơi
	AgentNote            string `json:"agent_note"`             // ghi chú tổn đài viên
	PlayDate             int64  `json:"play_date"`              // ngày chơi

	// thông tin hóa đơn
	PaymentType    string `json:"payment_type"`    // loại thanh toán
	Company        string `json:"company"`         // công ty
	CompanyAddress string `json:"company_address"` // địa chỉ công ty
	CompanyTax     string `json:"company_tax"`     // Mã số thuế
	ReceiptEmail   string `json:"receipt_email"`   // email nhận hóa đơn

	BookingList []models_agency_booking.AgencyBookingInfo `json:"booking_list"` // danh sách booking
}

type GetAgencyTransactionRequest struct {
	PageRequest
	TransactionId        string `json:"transaction_id"`
	CourseId             string `form:"course_uid" json:"course_uid" binding:"required"`
	PartnerUid           string `form:"partner_uid" json:"partner_uid" binding:"required"`
	AgencyId             string `json:"agency_id"`              // Mã đại lí
	PaymentStatus        string `json:"payment_status"`         // trạng thái thanh toán
	TransactionStatus    string `json:"transaction_status"`     // trạng thái đơn ~ trạng thái giao dịch
	BookingRequestStatus string `json:"booking_request_status"` // trạng thái yêu cầu booking với nhà cung cấp
	CustomerPhoneNumber  string `json:"customer_phone_number"`  // điện thoại người đặt
	CustomerEmail        string `json:"customer_email"`         // email người đặt
	CustomerName         string `json:"customer_name"`          // Tên người đặt
	BookingAmount        int64  `json:"booking_amount"`         // giá book
	ServiceAmount        int64  `json:"service_amount"`         // giá dịch vụ
	PaymentDueDate       int64  `json:"payment_due_date"`       // ngày hết hạn thanh toán
	FromDate             string `json:"from_date"`              // từ ngày
	ToDate               string `json:"to_date"`                // đến ngày
}
