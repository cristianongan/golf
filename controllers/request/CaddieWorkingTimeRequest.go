package request

type CaddieCheckInWorkingTimeBody struct {
	CaddieId string `json:"caddie_id" binding:"required"`
}

type CaddieCheckOutWorkingTimeBody struct {
	Id int64 `json:"id" binding:"required"`
}
type GetListCaddieWorkingTimeForm struct {
	PageRequest
	CaddieId   string `form:"caddie_id" json:"caddie_id"`
	CaddieName string `form:"caddie_name" json:"caddie_name"`
	From       int64  `form:"from"`
	To         int64  `form:"to"`
}

type UpdateCaddieWorkingTimeBody struct {
	CaddieId     *string `json:"caddie_id"`
	CheckInTime  *int64  `json:"checkin_time"`
	CheckOutTime *int64  `json:"checkout_time"`
}
