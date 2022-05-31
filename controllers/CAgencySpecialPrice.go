package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CAgencySpecialPrice struct{}

func (_ *CAgencySpecialPrice) CreateAgencySpecialPrice(c *gin.Context, prof models.CmsUser) {
	body := models.AgencySpecialPrice{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated() {
		response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	body.Input = prof.UserName
	errC := body.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, body)
}

func (_ *CAgencySpecialPrice) GetListAgencySpecialPrice(c *gin.Context, prof models.CmsUser) {
	form := request.GetListAgencySpecialPriceForm{}
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
	agencyR := models.AgencySpecialPrice{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	agencyR.AgencyId = form.AgencyId
	agencyR.Status = form.Status
	list, total, err := agencyR.FindList(page)
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

func (_ *CAgencySpecialPrice) UpdateAgencySpecialPrice(c *gin.Context, prof models.CmsUser) {
	agencyIdStr := c.Param("id")
	agencyId, err := strconv.ParseInt(agencyIdStr, 10, 64)
	if err != nil || agencyId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	agency := models.AgencySpecialPrice{}
	agency.Id = agencyId
	errF := agency.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.AgencySpecialPrice{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Dow != agency.Dow {
		if body.IsDuplicated() {
			response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
	}

	if body.Dow != "" {
		agency.Dow = body.Dow
	}
	if body.FromHour != "" {
		agency.FromHour = body.FromHour
	}
	if body.ToHour != "" {
		agency.ToHour = body.ToHour
	}

	if body.GreenFee > 0 {
		agency.GreenFee = body.GreenFee
	}

	if body.CaddieFee > 0 {
		agency.CaddieFee = body.CaddieFee
	}

	if body.BuggyFee > 0 {
		agency.BuggyFee = body.BuggyFee
	}

	agency.Note = body.Note

	errUdp := agency.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, agency)
}

func (_ *CAgencySpecialPrice) DeleteAgencySpecialPrice(c *gin.Context, prof models.CmsUser) {
	agencyIdStr := c.Param("id")
	agencyId, err := strconv.ParseInt(agencyIdStr, 10, 64)
	if err != nil || agencyId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	agency := models.AgencySpecialPrice{}
	agency.Id = agencyId
	errF := agency.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := agency.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
