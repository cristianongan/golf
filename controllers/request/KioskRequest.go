package request

type GetListKioskForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	KioskName  string `form:"kiosk_name"`
	Status     string `form:"status"`
	KioskType  string `form:"kiosk_type"`
}
type CreateKioskForm struct {
	PartnerUid  string `json:"partner_uid"`
	CourseUid   string `json:"course_uid"`
	KioskName   string `json:"kiosk_name"`
	ServiceType string `json:"service_type"`
	KioskType   string `json:"kiosk_type"`
	Status      string `json:"status"`
}

type GetSetupListForm struct {
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}
