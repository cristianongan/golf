package request

type CreateParOfHoleBody struct {
	PartnerUid string            `json:"partner_uid" binding:"required"` // Hãng Golf
	CourseUid  string            `json:"course_uid" binding:"required"`  // Sân Golf
	Configs    []ConfigParOfHole `json:"configs" binding:"required"`     // Loại sân
}

type ConfigParOfHole struct {
	Action string `json:"action" binding:"required"` // Loại sân
	Id     int64  `json:"id"`                        // Số hố
	Course string `json:"course"`                    //  Sân
	Hole   int    `json:"hole"`                      // Số hố
	Par    int    `json:"par"`                       // Số lần chạm gậy
	Minute int    `json:"minute"`                    // Số phút
}

type GetListParOfHoleForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"` // Sân Golf
	Status     string `form:"status"`
}

type UpdateParOfHoleBody struct {
	Status     string `json:"status"`
	CourseType string `json:"course_tye" binding:"required"` // Loại sân
	Course     string `json:"course" binding:"required"`     //  Sân
	Hole       int    `json:"hole"`                          // Số hố
	Par        int    `json:"par"`                           // Số lần chạm gậy
	Minute     int    `json:"minute"`                        // Số phút
}

type ResetParOfHoleBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"` // Sân Golf
	Course     string `json:"course" binding:"required"`     //  Sân
}
