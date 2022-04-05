package request

type GetListBookingSettingGroupForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}

type GetListBookingSettingForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	GroupId    int64  `form:"group_id"`
}

type GetListBookingForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}

type CreateBookingbody struct {
	PartnerUid     string `json:"partner_uid"`      // Hang Golf
	CourseUid      string `json:"course_uid"`       // San Golf
	Bag            string `json:"bag"`              // Golf Bag
	Hole           int    `json:"hole"`             // Số hố
	GuestStyle     string `json:"guest_style"`      // Guest Style
	GuestStyleName string `json:"guest_style_name"` // Guest Style Name
	CustomerName   string `json:"customer_name"`    // Tên khách hàng
	TeeType        string `json:"tee_type"`         // Tee1, Tee10, Tea1A, Tea1B, Tea1C
}
