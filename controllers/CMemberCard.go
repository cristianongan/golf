package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CMemberCard struct{}

func (_ *CMemberCard) CreateMemberCard(c *gin.Context, prof models.CmsUser) {
	body := models.MemberCard{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// Check Member Card Type Exit
	mcType := models.MemberCardType{}
	mcType.Id = body.McTypeId
	errFind := mcType.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	// Check Owner Invalid
	owner := models.CustomerUser{}
	owner.Uid = body.OwnerUid
	errFind = owner.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	// Check duplicated
	if body.IsDuplicated() {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	memberCard := models.MemberCard{
		CardId:   body.CardId,
		McTypeId: body.McTypeId,
	}

	memberCard.PartnerUid = body.PartnerUid
	memberCard.CourseUid = body.CourseUid

	memberCard.OwnerUid = body.OwnerUid
	memberCard.ValidDate = body.ValidDate
	memberCard.ExpDate = body.ExpDate
	memberCard.Note = body.Note
	memberCard.ReasonUnactive = body.ReasonUnactive
	memberCard.ChipCode = body.ChipCode

	memberCard.PriceCode = body.PriceCode
	memberCard.GreenFee = body.GreenFee
	memberCard.CaddieFee = body.CaddieFee
	memberCard.BuggyFee = body.BuggyFee
	memberCard.AdjustPlayCount = body.AdjustPlayCount
	memberCard.Float = body.Float

	errC := memberCard.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, memberCard)
}

func (_ *CMemberCard) GetListMemberCard(c *gin.Context, prof models.CmsUser) {
	form := request.GetListMemberCardForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	memberCardR := models.MemberCard{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		McTypeId:   form.McTypeId,
		OwnerUid:   form.OwnerUid,
		CardId:     form.CardId,
	}
	memberCardR.Status = form.Status
	list, total, err := memberCardR.FindList(page, form.PlayerName)
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

func (_ *CMemberCard) UpdateMemberCard(c *gin.Context, prof models.CmsUser) {
	memberCardUidStr := c.Param("uid")
	if memberCardUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	memberCard := models.MemberCard{}
	memberCard.Uid = memberCardUidStr
	errF := memberCard.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.MemberCard{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.OwnerUid != "" {
		memberCard.OwnerUid = body.OwnerUid
	}
	if body.Status != "" {
		memberCard.Status = body.Status
	}
	if body.ReasonUnactive != "" {
		memberCard.ReasonUnactive = body.ReasonUnactive
	}
	memberCard.PriceCode = body.PriceCode
	memberCard.GreenFee = body.GreenFee
	memberCard.CaddieFee = body.CaddieFee
	memberCard.BuggyFee = body.BuggyFee
	memberCard.Note = body.Note
	memberCard.ValidDate = body.ValidDate
	memberCard.StartPrecial = body.StartPrecial
	memberCard.EndPrecial = body.EndPrecial
	memberCard.AdjustPlayCount = body.AdjustPlayCount
	memberCard.Float = body.Float
	memberCard.PromotionCode = body.PromotionCode
	memberCard.UserEdit = body.UserEdit

	errUdp := memberCard.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, memberCard)
}

func (_ *CMemberCard) DeleteMemberCard(c *gin.Context, prof models.CmsUser) {
	memberCardUidStr := c.Param("uid")
	if memberCardUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	member := models.MemberCard{}
	member.Uid = memberCardUidStr
	errF := member.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := member.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

func (_ *CMemberCard) GetDetail(c *gin.Context, prof models.CmsUser) {
	memberCardUidStr := c.Param("uid")
	if memberCardUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	memberCard := models.MemberCard{}
	memberCard.Uid = memberCardUidStr
	errF := memberCard.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	memberDetailRes, errFind := memberCard.FindDetail()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	okResponse(c, memberDetailRes)
}

func (_ *CMemberCard) UnactiveMemberCard(c *gin.Context, prof models.CmsUser) {
	body := request.LockMemberCardBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	memberCard := models.MemberCard{}
	memberCard.Uid = body.MemberCardUid
	errF := memberCard.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	memberCard.Status = body.Status
	memberCard.ReasonUnactive = body.ReasonUnactive

	errUdp := memberCard.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, memberCard)
}
