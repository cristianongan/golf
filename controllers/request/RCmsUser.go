package request

type LoginBody struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Ttl      int    `json:"ttl"`
}

type GetListCmsUserForm struct {
	PageRequest
	CourseUid  string `form:"course_uid"`
	PartnerUid string `form:"partner_uid"`
}
