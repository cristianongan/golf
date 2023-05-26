package request

type CreateRestaurantSettingBody struct {
	PartnerUid    string   `json:"partner_uid" binding:"required"` // Hãng Golf
	CourseUid     string   `json:"course_uid" binding:"required"`  // Sân Golf
	ServiceIds    []int64  `json:"service_ids" binding:"required"`
	Name          string   `json:"name" binding:"required"`          // Tên setting
	NumberTables  int      `json:"number_tables" binding:"required"` // Số bàn
	PeopleInTable int      `json:"people_in_table"`                  //  Tổng số người trong 1 bàn
	Type          string   `json:"type"`                             // Loại setting
	Status        string   `json:"status"`                           // Loại setting
	Time          int      `json:"time"`                             // Số phút setting
	Symbol        string   `json:"symbol"`                           // Ký hiệu
	TableFrom     int      `json:"table_from"`                       //
	DataTables    []string `json:"data_tables"`                      //
}

type GetListRestaurantSettingForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" binding:"required"`
	CourseUid  string `form:"course_uid" binding:"required"` // Sân Golf
}

type UpdateRestaurantSettingBody struct {
	Name          string   `json:"name" binding:"required"`          // Tên setting
	NumberTables  int      `json:"number_tables" binding:"required"` // Số bàn
	PeopleInTable int      `json:"people_in_table"`                  //  Tổng số người trong 1 bàn
	Status        string   `json:"status"`                           // Loại setting
	Time          int      `json:"time"`                             // Số phút setting
	Symbol        string   `json:"symbol"`                           // Ký hiệu
	TableFrom     int      `json:"table_from"`                       //
	DataTables    []string `json:"data_tables"`
}
