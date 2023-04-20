package request

type EkycGetMemberCardList struct {
	CheckSum   string `json:"check_sum" binding:"required"`
	PartnerUid string `json:"partner_uid" binding:"required"`
	CourseUid  string `json:"course_uid" binding:"required"`
}

type EkycCheckBookingMember struct {
	CheckSum    string `json:"check_sum" binding:"required"`
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid" binding:"required"`
	MemberUid   string `json:"member_uid" binding:"required"`
	BookingDate string `json:"booking_date" binding:"required"` // dd/MM/yyyy
	MemberId    string `json:"member_id"`
}

type EkycCheckInBookingMember struct {
	CheckSum    string `json:"check_sum" binding:"required"`
	PartnerUid  string `json:"partner_uid" binding:"required"`
	CourseUid   string `json:"course_uid" binding:"required"`
	BookingUid  string `json:"booking_uid" binding:"required"`
	BookingDate string `json:"booking_date" binding:"required"` // dd/MM/yyyy
	MemberUid   string `json:"member_uid"`
}
