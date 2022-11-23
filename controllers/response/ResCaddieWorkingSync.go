package response

import (
	"start/models"
)

type CaddieWorkingSyncRes struct {
	Total int64               `json:"total"`
	Data  []CaddieWorkingSync `json:"data"`
}

type CaddieWorkingSync struct {
	models.ModelId
	PartnerUid string `json:"partner_uid"` // Hang Golf
	CourseUid  string `json:"course_uid"`  // San Golf
	EmployeeID string `json:"employee_id"` // Id caddie máy chấm công
	CheckIn    int64  `json:"check_in"`    // Thời gian check in
	CheckOut   int64  `json:"check_out"`   // Thời gian check out
	TotalTime  int    `json:"total_time"`  // Thời gian làm việc
	Date       string `json:"date"`        // Ngày làm việc
}
