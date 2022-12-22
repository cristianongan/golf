package request

type CreateCaddieBody struct {
	Code       string `json:"code"` // id caddie
	CourseUid  string `json:"course_uid"`
	PartnerUid string `json:"partner_uid"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Sex        bool   `json:"sex"`
	BirthDay   int64  `json:"birth_day"`
	//WorkingStatus string `json:"working_status"`
	Group          string `json:"group"`
	GroupId        int64  `json:"group_id"`
	StartedDate    int64  `json:"started_date"`
	CurrentStatus  string `json:"current_status"`
	IdHr           string `json:"id_hr"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	IdentityCard   string `json:"identity_card"`
	IssuedBy       string `json:"issued_by"`
	ExpiredDate    int64  `json:"expired_date"`
	PlaceOfOrigin  string `json:"place_of_origin"`
	Address        string `json:"address"`
	Level          string `json:"level"`
	ContractStatus string `json:"contract_status"`
	Note           string `json:"note"`
}

type GetListCaddieForm struct {
	PageRequest
	CourseId          string `form:"course_uid" json:"course_uid" binding:"required"`
	PartnerUid        string `form:"partner_uid" json:"partner_uid" binding:"required"`
	WorkingStatus     string `form:"working_status" json:"working_status"`
	Level             string `form:"level" json:"level"`
	Name              string `form:"name" json:"name"`
	Code              string `form:"code" json:"code"`
	Phone             string `form:"phone" json:"phone"`
	GroupId           string `form:"group_id"`
	IsInGroup         string `form:"is_in_group"`
	IsReadyForBooking string `form:"is_ready_for_booking"`
	ContractStatus    string `form:"contract_status"`
	CurrentStatus     string `form:"current_status"`
	IsReadyForJoin    string `form:"is_ready_for_join"`
	IsBooked          string `form:"is_booked"`
}

type GetListCaddieReady struct {
	CourseId   string `form:"course_uid" json:"course_uid" binding:"required"`
	PartnerUid string `form:"partner_uid" json:"partner_uid" binding:"required"`
	DateTime   string `form:"date_time"`
}

type UpdateCaddieBody struct {
	Code           string  `json:"code"` // id caddie
	CourseId       *string `json:"course_uid"`
	Name           *string `json:"name"`
	Avatar         *string `json:"avatar"`
	Sex            *bool   `json:"sex"`
	BirthDay       *int64  `json:"birth_day"`
	WorkingStatus  *string `json:"working_status"`
	Group          *string `json:"group"`
	StartedDate    *int64  `json:"started_date"`
	IdHr           *string `json:"id_hr"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	IdentityCard   *string `json:"identity_card"`
	IssuedBy       *string `json:"issued_by"`
	ExpiredDate    *int64  `json:"expired_date"`
	PlaceOfOrigin  *string `json:"place_of_origin"`
	Level          *string `json:"level"`
	Address        *string `json:"address"`
	Note           *string `json:"note"`
	GroupId        int64   `json:"group_id"`
	ContractStatus *string `json:"contract_status"`
}
