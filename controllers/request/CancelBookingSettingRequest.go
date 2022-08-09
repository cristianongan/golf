package request

type GetCancelBookingRequest struct {
	PartnerUid string `json:"partner_uid"`
	CourseUid  string `json:"course_uid"`
}
