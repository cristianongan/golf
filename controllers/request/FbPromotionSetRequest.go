package request

import (
	"start/utils"
)

type FbPromotionSetBody struct {
	PartnerUid  string           `json:"partner_uid" binding:"required"`
	CourseUid   string           `json:"course_uid" binding:"required"`
	GroupCode   string           `json:"group_code"`
	Code        string           `json:"code" binding:"required"`
	EnglishName string           `json:"english_name"`
	VieName     string           `json:"vietnamese_name"`
	Discount    int64            `json:"discount"`
	Note        string           `json:"note"`
	FBList      utils.ListString `json:"fb_list"`
	Status      string           `json:"status"`
	InputUser   string           `json:"input_user"`
	Price       float64          `json:"price"`
}

type GetListFbPromotionSetForm struct {
	PageRequest
	PartnerUid string `form:"partner_uid" json:"partner_uid"`
	CourseUid  string `form:"course_uid" json:"course_uid"`
	GroupCode  string `form:"group_code" json:"group_code"`
	Status     string `form:"status" json:"status"`
	CodeOrName string `form:"code_or_name"`
}

type UpdateFbPromotionSet struct {
	EnglishName string           `json:"english_name"`
	VieName     string           `json:"vietnamese_name"`
	SetName     *string          `json:"set_name"`
	Discount    *int64           `json:"discount"`
	Note        *string          `json:"note"`
	FBList      utils.ListString `json:"fb_list"`
	Status      *string          `form:"status" json:"status"`
	Price       float64          `json:"price"`
}
