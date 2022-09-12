package request

type AddRoundBody struct {
	BookUidList []string `json:"booking_uid_list" validate:"required"`
	Hole        *int     `json:"hole"`
}

type SplitRoundBody struct {
	BookingUid string `json:"booking_uid"`
	Hole       int    `json:"hole"`
	RoundId    int64  `json:"round_id"`
}

type MergeRoundBody struct {
	BookingUid string `json:"booking_uid"`
}
