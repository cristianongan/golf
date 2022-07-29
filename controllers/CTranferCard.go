package controllers

import (
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CTranferCard struct{}

func (_ *CTranferCard) CreateTranferCard(c *gin.Context, prof models.CmsUser) {
	body := request.CreateTranferCardBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	// Check Owner Old Invalid
	memberCard := models.MemberCard{}
	memberCard.Uid = body.CardUid
	errF := memberCard.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	// Check Owner Old Invalid
	ownerOld := models.CustomerUser{}
	ownerOld.Uid = body.OwnerOldUid
	errFind := ownerOld.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	// Check Owner New Invalid
	owner := models.CustomerUser{}
	owner.Uid = body.OwnerNewUid
	errFind = owner.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	tranferCard := models.TranferCard{}

	tranferCard.PartnerUid = body.PartnerUid
	tranferCard.CourseUid = body.CourseUid

	// Uid owner old
	tranferCard.OwnerUidOld = body.OwnerOldUid

	// Uid owner new
	tranferCard.OwnerUid = body.OwnerNewUid

	// update member card
	memberCard.OwnerUid = body.OwnerNewUid

	if body.ExpDate != 0 {
		memberCard.ExpDate = body.ExpDate
	}

	errUdp := memberCard.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	// Tranfer Card Infor
	tranferCard.CardUid = body.CardUid
	tranferCard.CardId = body.CardId
	tranferCard.TranferDate = body.TranferDate
	tranferCard.BillNumber = body.BillNumber
	tranferCard.BillDate = body.BillDate
	tranferCard.Amount = body.Amount
	tranferCard.InputUser = body.InputUser

	errC := tranferCard.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, tranferCard)
}

func (_ *CTranferCard) GetListTranferCard(c *gin.Context, prof models.CmsUser) {
	body := request.GetTranferCardList{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   body.PageRequest.Limit,
		Page:    body.PageRequest.Page,
		SortBy:  body.PageRequest.SortBy,
		SortDir: body.PageRequest.SortDir,
	}

	tranferCardR := models.TranferCard{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		CardId:     body.CardId,
	}

	list, total, err := tranferCardR.FindList(page, body.PlayerName)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": total,
		"data":  list,
	}

	okResponse(c, res)
}
