package request

type GetListCourseOTABody struct {
	PartnerUid string `json:"PartnerUid"`
}

type GetListTeeTypeInfoOTABody struct {
	PartnerUid string `json:"PartnerUid" binding:"required"`
	CourseCode string `json:"CourseCode" binding:"required"`
}
