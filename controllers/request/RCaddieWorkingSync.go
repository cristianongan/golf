package request

type GetDetalCaddieWorkingSyncBody struct {
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
	Week       int    `json:"week" binding:"required"`
	EmployeeId string `json:"employee_id" binding:"required"`
}
