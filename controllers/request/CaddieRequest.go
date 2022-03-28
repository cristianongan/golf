package request

type CreateCaddieBody struct {
	CourseId       string `json:"course_id" binding:"required"`
	Num            string `json:"caddie_num" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	Image          string `json:"image"`
	Sex            bool   `json:"sex"`
	BirthDay       int64  `json:"birth_day"`
	BirthPlace     string `json:"birth_place"`
	IdentityCard   string `json:"identity_card"`
	IssuedBy       string `json:"issued_by"`
	IssuedDate     int64  `json:"issued_date"`
	ExpiredDate    int64  `json:"expired_date"`
	EducationLevel string `json:"education_level"`
	FingerPrint    string `json:"finger_print"`
	HrCode         string `json:"hr_code"`
	HrPosition     string `json:"hr_position"`
	Group          string `json:"group"`
	StartedDate    int64  `json:"started_date"`
	WorkingStatus  string `json:"working_status"`
	Level          string `json:"level"`
	Note           string `json:"note"`
}

type GetListCaddieForm struct {
	PageRequest
	CourseId string `form:"course_id" json:"course_id"`
}

type UpdateCaddieBody struct {
	Num            *string `json:"caddie_num"`
	Name           *string `json:"name"`
	Phone          *string `json:"phone"`
	Address        *string `json:"address"`
	Image          *string `json:"image"`
	Sex            *bool   `json:"sex"`
	BirthDay       *int64  `json:"birth_day"`
	BirthPlace     *string `json:"birth_place"`
	IdentityCard   *string `json:"identity_card"`
	IssuedBy       *string `json:"issued_by"`
	IssuedDate     *int64  `json:"issued_date"`
	ExpiredDate    *int64  `json:"expired_date"`
	EducationLevel *string `json:"education_level"`
	FingerPrint    *string `json:"finger_print"`
	HrCode         *string `json:"hr_code"`
	HrPosition     *string `json:"hr_position"`
	Group          *string `json:"group"`
	StartedDate    *int64  `json:"started_date"`
	WorkingStatus  *string `json:"working_status"`
	Level          *string `json:"level"`
	Note           *string `json:"note"`
}
