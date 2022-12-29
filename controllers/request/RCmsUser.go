package request

type CreateCmsUserBody struct {
	UserName   string `json:"user_name" binding:"required"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid"`
	RoleId     int64  `json:"role_id"`
	Password   string `json:"password" binding:"required"`
}

type UdpCmsUserBody struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RoleId   int64  `json:"role_id"`
	Status   string `json:"status"`
}

type LoginBody struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Ttl      int    `json:"ttl"`
}

type GetListCmsUserForm struct {
	PageRequest
	CourseUid  string `form:"course_uid"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	Search     string `form:"search"`
}
