package request

type CreateTranferCardBody struct {
	PartnerUid  string `json:"partner_uid" validate:"required"`
	CourseUid   string `json:"course_uid"  validate:"required"`
	InputUser   string `json:"input_user" validate:"required"`
	OwnerOldUid string `json:"owner_old_uid" validate:"required"`
	OwnerNewUid string `json:"owner_new_uid" validate:"required"`
	CardUid     string `json:"card_uid" validate:"required"`
	CardId      string `json:"card_id" validate:"required"`
	Amount      int64  `json:"amount" validate:"required"`
	TranferDate int64  `json:"tranfer_date" validate:"required"`
	ExpDate     int64  `json:"exp_date"`
	BillNumber  string `json:"bill_number"`
	BillDate    int64  `json:"bill_date"`
	Note        string `json:"note"`
}

type GetTranferCardList struct {
	PageRequest
	PartnerUid string `form:"partner_uid"`
	CourseUid  string `form:"course_uid"`
	CardId     string `form:"card_id"`
	PlayerName string `form:"player_name"`
	OwnerId    string `form:"owner_id"`
}
