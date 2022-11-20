package request

type AddRoundBody struct {
	BookUidList []string `json:"booking_uid_list" validate:"required"`
	Hole        *int     `json:"hole"`
	CourseType  string   `json:"course_type"`
}

type SplitRoundBody struct {
	BookingUid string `json:"booking_uid"`
	Hole       int    `json:"hole"`
	RoundId    int64  `json:"round_id"`
}

type MergeRoundBody struct {
	BookingUid string `json:"booking_uid"`
}

type ChangeGuestyleRound struct {
	RoundId    int64  `json:"round_id"`
	BookingUid string `json:"booking_uid"`
	GuestStyle string `json:"guest_style"`
}
