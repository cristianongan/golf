package request

type UpdateBookingSource struct {
	BookingSourceName string `json:"booking_source_name"`
	AgencyId          string `json:"agency_id"`
	IsPart1TeeType    *bool  `json:"is_part1_tee_type"`
	IsPart2TeeType    *bool  `json:"is_part2_tee_type"`
	IsPart3TeeType    *bool  `json:"is_part3_tee_type"`
	NormalDay         *bool  `json:"normal_day"`
	Weekend           *bool  `json:"week_end"`
	NumberOfDays      int64  `json:"number_of_days"`
	Status            string `json:"status"`
}
type GetListBookingSource struct {
	PageRequest
	PartnerUid        string `json:"partner_uid"`
	CourseUid         string `json:"course_uid"`
	BookingSourceName string `form:"booking_source_name"`
	Status            string `form:"status"`
}
