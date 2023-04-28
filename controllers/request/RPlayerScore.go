package request

type CreatePlayerScoreBody struct {
	PartnerUid  string   `json:"partner_uid" binding:"required"` // Hãng Golf
	CourseUid   string   `json:"course_uid" binding:"required"`  // Sân Golf
	BookingDate string   `json:"booking_date"`                   //  Sân
	Course      string   `json:"course"`                         //  Sân
	Hole        int      `json:"hole"`                           // Số hố
	Par         int      `json:"par"`                            // Số lần chạm gậy
	TimeStart   int64    `json:"time_start"`
	TimeEnd     int64    `json:"time_end"`
	Players     []Player `json:"players"`
}

type Player struct {
	Bag   string `json:"bag"` //  Bag
	Shots int    `json:"shots"`
	Index int    `json:"index"`
}

type GetListPlayerScoreForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"` // Sân Golf
	BookingUid string `form:"booking_uid"`
}

type UpdatePlayerScoreBody struct {
	Course    string `json:"course"` //  Sân
	Hole      int    `json:"hole"`   // Số hố
	Par       int    `json:"par"`    // Số lần chạm gậy
	Shots     int    `json:"shots"`
	Index     int    `json:"index"`
	TimeStart int64  `json:"time_start"`
	TimeEnd   int64  `json:"time_end"`
}
