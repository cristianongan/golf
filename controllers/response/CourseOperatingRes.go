package response

import (
	model_gostarter "start/models/go-starter"
)

type CourseOperatingResponse struct {
	PartnerUid  string                `json:"partner_uid,omitempty"`
	CourseUid   string                `json:"course_uid,omitempty"`
	Course      string                `json:"course"` //  Sân
	Hole        int                   `json:"hole"`   // Số hố
	Minute      int                   `json:"minute"` // Số phút
	DataWaiting []model_gostarter.Map `json:"data_waiting,omitempty"`
	DataPlayed  []model_gostarter.Map `json:"data_played,omitempty"`
}
