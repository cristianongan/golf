package request

import "start/models"

type CreatePlayerScoreBody struct {
	PartnerUid  string   `json:"partner_uid" binding:"required"` // Hãng Golf
	CourseUid   string   `json:"course_uid" binding:"required"`  // Sân Golf
	BookingDate string   `json:"booking_date"`                   //  Sân
	Course      string   `json:"course"`                         //  Sân
	Hole        int      `json:"hole"`                           // Số hố
	HoleIndex   int      `json:"hole_index"`                     // Số thứ tự của hố
	Par         int      `json:"par"`                            // Số lần chạm gậy
	TimeStart   int64    `json:"time_start"`
	TimeEnd     int64    `json:"time_end"`
	FlightId    int64    `json:"flight_id" binding:"required"`
	Players     []Player `json:"players"`
}

type Player struct {
	Bag   string `json:"bag"` //  Bag
	Shots int    `json:"shots"`
	Index int    `json:"index"`
}

type GetListPlayerScoreForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid" binding:"required"`
	CourseUid   string `form:"course_uid" binding:"required"` // Sân Golf
	BookingDate string `form:"booking_date"`                  //  Sân
	Bag         string `form:"bag"`
	Hole        int    `form:"hole"`
	HoleIndex   int    `form:"hole_index"` // Số thứ tự của hố
	FlightId    int64  `form:"flight_id"`
	Status      string `form:"status"`
}

type UpdatePlayerScoreBody struct {
	Course    string `json:"course"` //  Sân
	Hole      int    `json:"hole"`   // Số hố
	Par       int    `json:"par"`    // Số lần chạm gậy
	Shots     int    `json:"shots"`
	Index     int    `json:"index"`
	TimeStart int64  `json:"time_start"`
	TimeEnd   int64  `json:"time_end"`
	FlightId  int64  `json:"flight_id"`
	HoleIndex int    `json:"hole_index"` // Số thứ tự của hố
}

type UpdateListPSBody struct {
	ListPlayer []models.PlayerScore `json:"list_player"` //  Sân
}
