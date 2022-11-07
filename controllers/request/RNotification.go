package request

import "start/models"

type CreateNotificationBody struct {
	Name   string `json:"name" binding:"required"`
	Uid    string `json:"uid" binding:"required"`
	Status string `json:"status"`
}

type GetListNotificationForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
}

type GetCaddieVacationNotification struct {
	Caddie       models.Caddie
	DateFrom     int64
	DateTo       int64
	NumberDayOff int
	Title        string
	CreateAt     int64
	UserName     string
}

type UpdateNotificationBody struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
