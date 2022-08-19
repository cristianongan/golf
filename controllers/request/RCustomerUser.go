package request

import "start/utils"

type GetListCustomerUserForm struct {
	PageRequest
	PartnerUid  string `form:"partner_uid"`
	CourseUid   string `form:"course_uid"`
	Type        string `form:"type"`
	CustomerUid string `form:"customer_uid"`
	Name        string `form:"name"`
	AgencyId    int64  `form:"agency_id"`
	Phone       string `form:"phone"`
	Identify    string `form:"identify"`
}

type DeleteAgencyCustomerUser struct {
	CusUserUids utils.ListString `json:"cus_user_uids"`
}

type GetBirthdayList struct {
	PageRequest
	FromDate string `form:"from_date"`
	ToDate   string `form:"to_date"`
}
