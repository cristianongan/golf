package request

type CreateTeeTypeClose struct {
	PartnerUid       string `json:"partner_uid" binding:"required"`
	CourseUid        string `json:"course_uid" binding:"required"`
	BookingSettingId int64  `json:"booking_setting_id" binding:"required"`
	DateTime         string `json:"date_time" binding:"required"`
	Note             string `json:"note"`
}

type GetListTeeTypeClose struct {
	PageRequest
	PartnerUid       string `form:"partner_uid"`
	CourseUid        string `form:"course_uid"`
	BookingSettingId *int64 `json:"booking_setting_id"`
	DateTime         string `form:"date_time"`
}
