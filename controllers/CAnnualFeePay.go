package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CAnnualFeePay struct{}

func (_ *CAnnualFeePay) CreateAnnualFeePay(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.AnnualFeePay{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// Check member card exits
	memberCard := models.MemberCard{}
	memberCard.Uid = body.MemberCardUid
	errFind := memberCard.FindFirst(db)
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	annualFeePay := models.AnnualFeePay{
		PartnerUid:    body.PartnerUid,
		CourseUid:     body.CourseUid,
		Year:          body.Year,
		MemberCardUid: body.MemberCardUid,
		PaymentType:   body.PaymentType,
		BillNumber:    body.BillNumber,
		Note:          body.Note,
		Amount:        body.Amount,
		PayDate:       body.PayDate,
		CmsUser:       prof.UserName,
	}

	errC := annualFeePay.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	//Update total paid của membercard trong năm
	go updateTotalPaidAnnualFeeForMemberCard(db, body.MemberCardUid, body.Year)
	go updateReportTotalPaidForCustomerUser(db, memberCard.OwnerUid, body.PartnerUid, body.CourseUid)

	okResponse(c, annualFeePay)
}

func (_ *CAnnualFeePay) GetListAnnualFeePay(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListAnnualFeeForm{}
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

	annualFeeR := models.AnnualFeePay{
		PartnerUid:    form.PartnerUid,
		CourseUid:     form.CourseUid,
		MemberCardUid: form.MemberCardUid,
		Year:          form.Year,
	}
	list, total, err := annualFeeR.FindList(db, page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//get total paid
	totalPaid := annualFeeR.FindTotalPaid(db)

	res := map[string]interface{}{
		"total":      total,
		"data":       list,
		"total_paid": totalPaid,
	}

	okResponse(c, res)
}

func (_ *CAnnualFeePay) UpdateAnnualFeePay(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	annualFeeIdStr := c.Param("id")
	annualFeeId, err := strconv.ParseInt(annualFeeIdStr, 10, 64)
	if err != nil || annualFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	annualFeePay := models.AnnualFeePay{}
	annualFeePay.Id = annualFeeId
	errF := annualFeePay.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.AnnualFeePay{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	annualFeePay.Amount = body.Amount
	annualFeePay.CmsUser = prof.UserName

	errUdp := annualFeePay.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	//Update total paid của membercard trong năm
	go updateTotalPaidAnnualFeeForMemberCard(db, body.MemberCardUid, body.Year)

	memberCard := models.MemberCard{}
	memberCard.Uid = annualFeePay.MemberCardUid
	errFMC := memberCard.FindFirst(db)
	if errFMC == nil && memberCard.OwnerUid != "" {
		go updateReportTotalPaidForCustomerUser(db, memberCard.OwnerUid, memberCard.PartnerUid, memberCard.CourseUid)
	}

	okResponse(c, annualFeePay)
}

func (_ *CAnnualFeePay) DeleteAnnualFeePay(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	annualFeeIdStr := c.Param("id")
	annualFeeId, err := strconv.ParseInt(annualFeeIdStr, 10, 64)
	if err != nil || annualFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	annualFeePay := models.AnnualFeePay{}
	annualFeePay.Id = annualFeeId
	errF := annualFeePay.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := annualFeePay.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	//Update total paid của membercard trong năm
	go updateTotalPaidAnnualFeeForMemberCard(db, annualFeePay.MemberCardUid, annualFeePay.Year)
	memberCard := models.MemberCard{}
	memberCard.Uid = annualFeePay.MemberCardUid
	errFMC := memberCard.FindFirst(db)
	if errFMC == nil && memberCard.OwnerUid != "" {
		go updateReportTotalPaidForCustomerUser(db, memberCard.OwnerUid, memberCard.PartnerUid, memberCard.CourseUid)
	}

	okRes(c)
}
