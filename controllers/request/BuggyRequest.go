package request

type CreateBuggyBody struct {
	Code            string  `json:"code"` // id Buggy
	CourseUid       string  `json:"course_uid" binding:"required"`
	PartnerUid      string  `json:"partner_uid"`
	Origin          string  `json:"origin"`
	Note            string  `json:"note"`
	BuggyForVip     bool    `json:"buggy_for_vip"`
	WarrantyPeriod  float64 `json:"warranty_period"`
	MaintenanceFrom int64   `json:"maintenance_from"`
	MaintenanceTo   int64   `json:"maintenance_to"`
	BuggyStatus     string  `json:"buggy_status"`
}

type GetListBuggyForm struct {
	PageRequest
	Code        *string `form:"buggy_uid" json:"buggy_uid"`
	BuggyStatus *string `form:"buggy_status" json:"buggy_status"`
	BuggyForVip *bool   `form:"buggy_for_vip" json:"buggy_for_vip"`
}

type UpdateBuggyBody struct {
	Origin *string `json:"origin"`
	Note   *string `json:"note"`
}
