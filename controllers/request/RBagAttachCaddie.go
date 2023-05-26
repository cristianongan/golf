package request

type CreateBagAttachCaddieBody struct {
	PartnerUid   string `json:"partner_uid" validate:"required"`
	CourseUid    string `json:"course_uid"  validate:"required"`
	BookingUid   string `json:"booking_uid"`
	BookingDate  string `json:"booking_date" validate:"required"`
	CaddieCode   string `json:"caddie_code" validate:"required"`
	CustomerName string `json:"customer_name"`
	Bag          string `json:"bag" validate:"required"`
	LockerNo     string `json:"locker_no"`
}

type UpdateBagAttachCaddieBody struct {
	BookingUid   string `json:"booking_uid" validate:"required"`
	BookingDate  string `json:"booking_date" validate:"required"`
	CaddieCode   string `json:"caddie_code" validate:"required"`
	CustomerName string `json:"customer_name"`
	LockerNo     string `json:"locker_no"`
	Bag          string `json:"bag" validate:"required"`
}

type GetListAttachCaddieForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"` // Sân Golf
	BookingDate string `form:"booking_date"`
	Search      string `form:"search"`
	Bag         string `form:"bag"`
	CmsUser     string `form:"cms_user"`
}
