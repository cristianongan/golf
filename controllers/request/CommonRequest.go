package request

type CommonRequest struct {
	PartnerUid string `json:"partner_uid" form:"partner_uid"`
}
