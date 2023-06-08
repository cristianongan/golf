package response

type GetTeeTimeAgencyResponse struct {
	Data         []TeeTimeOTA `json:"data"`
	IsMainCourse bool         `json:"is_main_course"`
	Token        interface{}  `json:"token"`
	CourseUid    string       `json:"course_uid"`
	AgencyId     string       `json:"agency_id"`
	Date         string       `json:"date"`
	NumTeeTime   int64        `json:"num_tee_time"`
}
