package request

type CreateCaddieBody struct {
	CaddieId      string `json:"caddie_id"`
	CourseId      string `json:"course_id"`
	Name          string `json:"name"`
	Sex           bool   `json:"sex"`
	BirthDay      int64  `json:"birth_day"`
	WorkingStatus string `json:"working_status"`
	Group         string `json:"group"`
	StartedDate   int64  `json:"started_date"`
	IdHr          string `json:"id_hr"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	IdentityCard  string `json:"identity_card"`
	IssuedBy      string `json:"issued_by"`
	ExpiredDate   int64  `json:"expired_date"`
	PlaceOfOrigin string `json:"place_of_origin"`
	Address       string `json:"address"`
	Level         string `json:"level"`
	Note          string `json:"note"`
}

type GetListCaddieForm struct {
	PageRequest
	CourseId string `form:"course_id" json:"course_id"`
}

type UpdateCaddieBody struct {
	CaddieId      *string `json:"caddie_num"`
	CourseId      *string `json:"course_id"`
	Name          *string `json:"name"`
	Sex           *bool   `json:"sex"`
	BirthDay      *int64  `json:"birth_day"`
	WorkingStatus *string `json:"working_status"`
	Group         *string `json:"group"`
	StartedDate   *int64  `json:"started_date"`
	IdHr          *string `json:"id_hr"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	IdentityCard  *string `json:"identity_card"`
	IssuedBy      *string `json:"issued_by"`
	ExpiredDate   *int64  `json:"expired_date"`
	PlaceOfOrigin *string `json:"place_of_origin"`
	Level         *string `json:"level"`
	Address       *string `json:"address"`
	Note          *string `json:"note"`
}
