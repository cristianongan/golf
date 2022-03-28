package controllers

import (
	"errors"
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

	memberCard := models.MemberCard{
		CardId:   body.CardId,
		McTypeId: body.McTypeId,
	}

	//Check Exits
	errFind := memberCard.FindFirst()
	if errFind == nil || memberCard.Uid != "" {
		response_message.DuplicateRecord(c, errors.New("Duplicate uid").Error())
		return
	}

	memberCard.PartnerUid = body.PartnerUid
	memberCard.CourseUid = body.CourseUid

	memberCard.OwnerUid = body.OwnerUid
	memberCard.ValidDate = body.ValidDate
	memberCard.ExpDate = body.ExpDate
	memberCard.Note = body.Note
	memberCard.ChipCode = body.ChipCode

	memberCard.PriceCode = body.PriceCode
	memberCard.GreenFee = body.GreenFee
	memberCard.CaddieFee = body.CaddieFee
	memberCard.BuggyFee = body.BuggyFee
	memberCard.AdjustPlayCount = body.AdjustPlayCount

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
	}
	list, total, err := memberCardR.FindList(page)
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
