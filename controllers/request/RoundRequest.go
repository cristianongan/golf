package request

type AddRoundBody struct {
	BookUidList []string `json:"booking_uid_list" validate:"required"`
	Hole        *int     `json:"hole"`
}

type SplitRoundBody struct {
	BookingUid string `json:"booking_uid"`
	Hole       int    `json:"hole"`
	RoundIndex int    `json:"round_index"`
}

type MergeRoundBody struct {
	BookingUid string `json:"booking_uid"`
}
