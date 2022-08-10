package request

type GetListMemberCardTypeForm struct {
	PageRequest
	Status     string `form:"status"`
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Name       string `form:"name"`
	Type       string `form:"type"`
	GuestStyle string `form:"guest_style"`
}

type GetFeeByHoleForm struct {
	McTypeId int64 `form:"mc_type_id" binding:"required"`
	Hole     int   `form:"hole"`
}

type AddMcAnnualFeeBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"` // Hang Golf
	CourseUid  string `json:"course_uid" binding:"required"`  // San Golf
	McTypeId   int64  `json:"mc_type_id" binding:"required"`  // Member Card Type id
	Year       int    `json:"year" binding:"required"`
	Fee        int64  `json:"fee" binding:"required"`
}

type UdpMcAnnualFeeBody struct {
	Id  int64 `json:"id" binding:"required"` // Id
	Fee int64 `json:"fee"`
}

type GetMcTypeAnnualFeeForm struct {
	McTypeId int64 `form:"mc_type_id" binding:"required"`
}
