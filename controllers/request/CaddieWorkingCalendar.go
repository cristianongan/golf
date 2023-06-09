package request

type CreateCaddieWorkingCalendarBody struct {
	PartnerUid        string                             `json:"partner_uid"`
	CourseUid         string                             `json:"course_uid"`
	CaddieWorkingList []CaddieWorkingCalendarListRequest `json:"caddie_working_list" binding:"required"`
	ActionType        string                             `json:"action_type"`
}

type ImportCaddieSlotAutoBody struct {
	PartnerUid string   `json:"partner_uid" binding:"required"`
	CourseUid  string   `json:"course_uid" binding:"required"`
	CaddieSlot []string `json:"caddie_slot" binding:"required"`
	ApplyDate  string   `json:"apply_date" binding:"required"`
}

type GetCaddieWorkingCalendarList struct {
	PartnerUid string `form:"partner_uid" validate:"required"`
	CourseUid  string `form:"course_uid" validate:"required"`
	ApplyDate  string `form:"apply_date" validate:"required"`
}

type GetNoteCaddieSlotByDateForm struct {
	PartnerUid string `form:"partner_uid" validate:"required"`
	CourseUid  string `form:"course_uid" validate:"required"`
	ApplyDate  string `form:"apply_date" validate:"required"`
}

type AddNoteCaddieSlotByDateForm struct {
	PartnerUid string `form:"partner_uid" validate:"required"`
	CourseUid  string `form:"course_uid" validate:"required"`
	ApplyDate  string `form:"apply_date" validate:"required"`
	Note       string `form:"note"`
}

type UpdateNoteCaddieSlotByDateForm struct {
	Note string `form:"note"`
}

type UpdateCaddieWorkingCalendarBody struct {
	CaddieCode string `json:"caddie_code" validate:"required"`
}

type UpdateCaddieWorkingSlotAutoBody struct {
	PartnerUid    string `json:"partner_uid" validate:"required"`
	CourseUid     string `json:"course_uid" validate:"required"`
	CaddieCodeOld string `json:"caddie_code_old" validate:"required"`
	CaddieCodeNew string `json:"caddie_code_new" validate:"required"`
	ApplyDate     string `json:"apply_date" validate:"required"`
}

type CaddieWorkingCalendarListRequest struct {
	ApplyDate  string                         `json:"apply_date"`
	Note       string                         `json:"note"`
	CaddieList []CaddieWorkingCalendarRequest `json:"caddie_list"`
}

type CaddieWorkingCalendarRequest struct {
	CaddieCode     string `json:"caddie_code"`
	Row            int    `json:"row"`
	NumberOrder    int64  `json:"number_order"`
	CaddieIncrease bool   `json:"caddie_increase"`
}
