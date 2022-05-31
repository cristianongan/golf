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

type CAgency struct{}

func (_ *CAgency) CreateAgency(c *gin.Context, prof models.CmsUser) {
	body := models.Agency{}
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

	errC := body.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, body)
}

func (_ *CAgency) GetListAgency(c *gin.Context, prof models.CmsUser) {
	form := request.GetListAgencyForm{}
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
	agencyR := models.Agency{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		Name:       form.Name,
		AgencyId:   form.AgencyId,
	}
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

func (_ *CAgency) UpdateAgency(c *gin.Context, prof models.CmsUser) {
	agencyIdStr := c.Param("id")
	agencyId, err := strconv.ParseInt(agencyIdStr, 10, 64)
	if err != nil || agencyId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	agency := models.Agency{}
	agency.Id = agencyId
	errF := agency.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.Agency{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if agency.AgencyId != body.AgencyId || agency.ShortName != body.ShortName {
		if body.IsDuplicated() {
			response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
	}

	if body.Name != "" {
		agency.Name = body.Name
	}
	if body.ShortName != "" {
		agency.ShortName = body.ShortName
	}
	if body.Category != "" {
		agency.Category = body.Category
	}
	if body.GuestStyle != "" {
		agency.GuestStyle = body.GuestStyle
	}
	if body.Province != "" {
		agency.Province = body.Province
	}
	if body.Status != "" {
		agency.Status = body.Status
	}
	agency.PrimaryContactFirst = body.PrimaryContactFirst
	agency.PrimaryContactSecond = body.PrimaryContactSecond
	agency.ContractDetail = body.ContractDetail

	errUdp := agency.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, agency)
}

func (_ *CAgency) DeleteAgency(c *gin.Context, prof models.CmsUser) {
	agencyIdStr := c.Param("id")
	agencyId, err := strconv.ParseInt(agencyIdStr, 10, 64)
	if err != nil || agencyId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	agency := models.Agency{}
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
