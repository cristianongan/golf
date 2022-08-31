package request

type KioskInventoryInputItemBody struct {
	PartnerUid    string `json:"partner_uid"`
	CourseUid     string `json:"course_uid"`
	Code          string `json:"code"`
	ItemCode      string `json:"item_code"`
	Quantity      int64  `json:"quantity"`
	Source        string `json:"source"`
	ReviewUserUid string `json:"review_user_uid"`
	Note          string `json:"note"`
	KioskCode     string `json:"kiosk_code"`
	KioskName     string `json:"kiosk_name"`
	KioskType     string `json:"kiosk_type"`
}

type KioskInventoryOutputItemBody struct {
	PartnerUid    string `json:"partner_uid"`
	CourseUid     string `json:"course_uid"`
	Code          string `json:"code"`
	ItemCode      string `json:"item_code"`
	Quantity      int64  `json:"quantity"`
	ReviewUserUid string `json:"review_user_uid"`
	Note          string `json:"note"`
	KioskCode     string `json:"kiosk_code"`
}

type KioskInventoryCreateItemBody struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type GetInputItems struct {
	PageRequest
	BillStatus string `json:"bill_status"`
	KioskName  string `json:"kiosk_name"`
	KioskType  string `json:"kiosk_type"`
}

type GetBillInput struct {
	PageRequest
	BillStatus string `json:"bill_status"`
	Type       string `json:"type"`
}

type KioskInventoryInsertBody struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	Code       string `json:"code"` // Mã đơn nhập
}
