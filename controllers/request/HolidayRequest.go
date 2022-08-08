package request

type GetListHolidayForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	Year       string `form:"year"`
}
type CreateHolidayForm struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
	Note       string `json:"note"`
	Name       string `json:"name"`
	From       string `json:"from"`
	To         string `json:"to"`
	Year       string `json:"year"`
}
