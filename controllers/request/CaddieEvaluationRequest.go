package request

type CreateCaddieEvaluationBody struct {
	BookingUid string `json:"booking_uid" validate:"required"`
	CaddieUid  string `json:"caddie_uid" validate:"required"`
	CaddieCode string `json:"caddie_code" validate:"required"`
	RankType   int    `json:"rank_type" validate:"required"`
}

type GetCaddieEvaluationList struct {
	PageRequest
	CaddieName string `form:"caddie_name"`
	CaddieCode string `form:"caddie_code"`
	Month      string `form:"month"`
}

type UpdateCaddieEvaluationBody struct {
	BookingUid string `json:"booking_uid" validate:"required"`
	CaddieUid  string `json:"caddie_uid" validate:"required"`
	CaddieCode string `json:"caddie_code" validate:"required"`
	RankType   int    `json:"rank_type"`
}
