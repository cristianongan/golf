package request

type CreateCaddieBody struct {
	Code          string `json:"code"` // id caddie
	CourseUid     string `json:"course_uid"`
	PartnerUid    string `json:"partner_uid"`
	Name          string `json:"name"`
	Avatar        string `json:"avatar"`
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
	CourseId string `form:"course_uid" json:"course_uid"`
}

type UpdateCaddieBody struct {
	Code          string  `json:"code"` // id caddie
	CourseId      *string `json:"course_uid"`
	Name          *string `json:"name"`
	Avatar        *string `json:"avatar"`
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
	IsInCourse    *bool   `json:"is_in_course"`
}
