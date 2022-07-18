package request

type CreateTeeTimeSettings struct {
	PartnerUid    string `json:"partner_uid" binding:"required"`
	CourseUid     string `json:"course_uid" binding:"required"`
	TeeTime       string `json:"tee_time" binding:"required"`
	TeeTimeStatus string `json:"tee_time_status" binding:"required"`
	DateTime      string `json:"date_time" binding:"required"`
	Note          string `json:"note"`
}
type GetListTeeTimeSettings struct {
	PageRequest
	PartnerUid    string `form:"partner_uid"`
	CourseUid     string `form:"course_uid"`
	TeeTime       string `form:"tee_time"`
	TeeTimeStatus string `form:"tee_time_status"`
	DateTime      string `form:"date_time"`
}
