package controllers

import (
	"errors"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CMemberCardType struct{}

func (_ *CMemberCardType) CreateMemberCardType(c *gin.Context, prof models.CmsUser) {
	body := models.MemberCardType{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	memberCardType := models.MemberCardType{
		Name: body.Name,
	}

	memberCardType.PartnerUid = body.PartnerUid
	memberCardType.CourseUid = body.CourseUid

	memberCardType.GuestStyle = body.GuestStyle
	memberCardType.GuestStyleOffGuest = body.GuestStyleOffGuest
	memberCardType.PromotGuestStyle = body.PromotGuestStyle
	memberCardType.NormalDayTakeGuest = body.NormalDayTakeGuest
	memberCardType.WeekendTakeGuest = body.WeekendTakeGuest
	memberCardType.PlayTimesOnMonth = body.PlayTimesOnMonth
	memberCardType.Type = body.Type

	errC := memberCardType.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, memberCardType)
}

func (_ *CMemberCardType) GetListMemberCardType(c *gin.Context, prof models.CmsUser) {
	form := request.GetListMemberCardTypeForm{}
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

	memberCardTypeR := models.MemberCardType{}
	list, total, err := memberCardTypeR.FindList(page)
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

func (_ *CMemberCardType) UpdateMemberCardType(c *gin.Context, prof models.CmsUser) {
	memberCardTypeIdStr := c.Param("id")
	memberCardTypeId, err := strconv.ParseInt(memberCardTypeIdStr, 10, 64)
	if err != nil || memberCardTypeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	memberCardType := models.MemberCardType{}
	memberCardType.Id = memberCardTypeId
	errF := memberCardType.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.MemberCardType{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		memberCardType.Name = body.Name
	}
	memberCardType.GuestStyle = body.GuestStyle
	memberCardType.GuestStyleOffGuest = body.GuestStyleOffGuest
	memberCardType.PromotGuestStyle = body.PromotGuestStyle
	memberCardType.NormalDayTakeGuest = body.NormalDayTakeGuest
	memberCardType.WeekendTakeGuest = body.WeekendTakeGuest
	memberCardType.PlayTimesOnMonth = body.PlayTimesOnMonth
	memberCardType.Status = body.Status
	memberCardType.Type = body.Type

	errUdp := memberCardType.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, memberCardType)
}

func (_ *CMemberCardType) DeleteMemberCardType(c *gin.Context, prof models.CmsUser) {
	memberCardTypeIdStr := c.Param("id")
	memberCardTypeId, err := strconv.ParseInt(memberCardTypeIdStr, 10, 64)
	if err != nil || memberCardTypeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	memberCardType := models.MemberCardType{}
	memberCardType.Id = memberCardTypeId
	errF := memberCardType.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := memberCardType.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
