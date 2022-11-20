package controllers

import (
	"errors"
	"start/constants"
	"start/models"
)

/*
 check partner permission
 isOpenVNPay = true: chức năng này VNPay dc quyền với tất cả các hãng
*/
func checkPermissionPartner(prof models.CmsUser, partnerBody string, isOpenVNPay bool) error {
	if prof.PartnerUid == constants.ROOT_PARTNER_UID {
		if isOpenVNPay {
			return nil
		}
	}

	if prof.PartnerUid == partnerBody {
		return nil
	}

	return errors.New("Not Permission")
}
