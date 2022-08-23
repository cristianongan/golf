package request

type AddRoundBody struct {
	BookUidList []string `json:"booking_uid_list" validate:"required"`
	Hole        int      `json:"hole" validate:"required"`
}

type SplitRoundBody struct {
	BookingUid string `json:"booking_uid"`
	Hole       int64  `json:"hole"`
	RoundIndex int64  `json:"round_index"`
}

type MergeRoundBody struct {
	BookingUid string `json:"booking_uid"`
}
