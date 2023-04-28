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
	CaddieId   int64  `json:"caddie_id"`
}

type UdpCmsUserBody struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	RoleId   int64  `json:"role_id"`
	Status   string `json:"status"`
	CaddieId int64  `json:"caddie_id"`
}

type LoginBody struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Type     string `json:"type"`
	Ttl      int    `json:"ttl"`
}

type GetListCmsUserForm struct {
	PageRequest
	CourseUid  string `form:"course_uid"`
	PartnerUid string `form:"partner_uid" binding:"required"`
	Search     string `form:"search"`
}

type ChangePassCmsUserBody struct {
	// UserUid string `json:"user_uid"  binding:"required"`
	OldPass string `json:"old_pass"  binding:"required"`
	NewPass string `json:"new_pass"  binding:"required"`
}

type ResetPassCmsUserBody struct {
	UserUid string `json:"user_uid"  binding:"required"`
}
