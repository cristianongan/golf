package request

type AddRoundBody struct {
	BookingUid string `json:"booking_uid"`
}

type SplitRoundBody struct {
	BookingUid string `json:"booking_uid"`
	Hole       int64  `json:"hole"`
	RoundIndex int64  `json:"round_index"`
}

type MergeRoundBody struct {
	BookingUid string `json:"booking_uid"`
}
