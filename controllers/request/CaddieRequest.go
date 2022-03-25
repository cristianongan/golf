package request

type CreateCaddieBody struct {
	CourseId       string `json:"course_id" binding:"required"`
	Num            string `json:"num" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	Sex            bool   `json:"sex"`
	BirthDay       int64  `json:"birth_day"`
	BirthPlace     string `json:"birth_place"`
	IdentityCard   string `json:"identity_card"`
	IssuedBy       string `json:"issued_by"`
	IssuedDate     int64  `json:"issued_date"`
	EducationLevel string `json:"education_level"`
	FingerPrint    string `json:"finger_print"`
	HrCode         string `json:"hr_code"`
	HrPosition     string `json:"hr_position"`
	Group          string `json:"group"`
	Row            string `json:"row"`
	StartedDate    int64  `json:"started_date"`
	RaisingChild   bool   `json:"raising_child"`
	TempAbsent     bool   `json:"temp_absent"`
	FullTime       bool   `json:"full_time"`
	WEWork         bool   `json:"we_work"`
	Level          string `json:"level"`
	Note           string `json:"note"`
}

type GetListCaddieForm struct {
	PageRequest
	CourseId string `form:"course_id" json:"course_id"`
}

type UpdateCaddieBody struct {
	Num            *string `json:"num"`
	Name           *string `json:"name"`
	Phone          *string `json:"phone"`
	Address        *string `json:"address"`
	Sex            *bool   `json:"sex"`
	BirthDay       *int64  `json:"birth_day"`
	BirthPlace     *string `json:"birth_place"`
	IdentityCard   *string `json:"identity_card"`
	IssuedBy       *string `json:"issued_by"`
	IssuedDate     *int64  `json:"issued_date"`
	EducationLevel *string `json:"education_level"`
	FingerPrint    *string `json:"finger_print"`
	HrCode         *string `json:"hr_code"`
	HrPosition     *string `json:"hr_position"`
	Group          *string `json:"group"`
	Row            *string `json:"row"`
	StartedDate    *int64  `json:"started_date"`
	RaisingChild   *bool   `json:"raising_child"`
	TempAbsent     *bool   `json:"temp_absent"`
	FullTime       *bool   `json:"full_time"`
	WEWork         *bool   `json:"we_work"`
	Level          *string `json:"level"`
	Note           *string `json:"note"`
}
