package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CAgency struct{}

func (_ *CAgency) CreateAgency(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.Agency{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated(db) {
		response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	if errContractNo := body.IsDuplicatedContract(db, body.ContractDetail.ContractNo); errContractNo == nil {
		response_message.BadRequest(c, "Contract No đã tồn tại!")
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
		Type:       form.Type,
	}
	agencyR.Status = form.Status
	list, total, err := agencyR.FindList(db, page)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	agencyIdStr := c.Param("id")
	agencyId, err := strconv.ParseInt(agencyIdStr, 10, 64)
	if err != nil || agencyId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	agency := models.Agency{}
	agency.Id = agencyId
	errF := agency.FindFirst(db)
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
		if body.IsDuplicated(db) {
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
	if body.Type != "" {
		agency.Type = body.Type
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

	errUdp := agency.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, agency)
}

func (_ *CAgency) DeleteAgency(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	agencyIdStr := c.Param("id")
	agencyId, err := strconv.ParseInt(agencyIdStr, 10, 64)
	if err != nil || agencyId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	agency := models.Agency{}
	agency.Id = agencyId
	errF := agency.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := agency.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

func (_ *CAgency) GetAgencyDetail(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	agencyIdStr := c.Param("id")
	agencyId, err := strconv.ParseInt(agencyIdStr, 10, 64)
	if err != nil || agencyId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	agency := models.Agency{}
	agency.Id = agencyId
	errF := agency.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	agencyDetail := models.AgencyDetailRes{
		Agency: agency,
	}
	//Get number customer
	numberCustomer := agency.GetNumberCustomer(db)
	agencyDetail.NumberOfCustomer = numberCustomer

	okResponse(c, agencyDetail)
}

/*
	Get base other price
*/
func (_ *CAgency) GetOtherBasePrice(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetOtherBasePriceForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.Type == constants.OTHER_BASE_PRICE_AGENCY && form.Id > 0 {
		// Get cho other price cho agency
		agencyPriceR := models.AgencySpecialPrice{
			AgencyId: form.Id,
		}

		agenPrice, errF := agencyPriceR.FindOtherPriceOnTime(db)
		if errF != nil {
			log.Println("GetOtherBasePrice errF", errF.Error())
		}

		res := map[string]interface{}{
			"green_fee":  agenPrice.GreenFee,
			"caddie_fee": agenPrice.CaddieFee,
			"buggy_fee":  agenPrice.BuggyFee,
		}
		okResponse(c, res)
		return
	} else if form.Type == constants.OTHER_BASE_PRICE_MEMBER_CARD && form.Uid != "" {
		memberCard := models.MemberCard{}
		memberCard.Uid = form.Uid
		errF := memberCard.FindFirst(db)
		if errF == nil {
			if memberCard.IsValidTimePrecial() {
				res := map[string]interface{}{
					"green_fee":  memberCard.GreenFee,
					"caddie_fee": memberCard.CaddieFee,
					"buggy_fee":  memberCard.BuggyFee,
				}
				okResponse(c, res)
				return
			}
		}
	}

	res := map[string]interface{}{
		"green_fee":  0,
		"caddie_fee": 0,
		"buggy_fee":  0,
	}

	okResponse(c, res)
}
