package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CAnnualFee struct{}

func (_ *CAnnualFee) CreateAnnualFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.AnnualFee{}
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

	// Check member card exits
	memberCard := models.MemberCard{}
	memberCard.Uid = body.MemberCardUid
	errFind := memberCard.FindFirst(db)
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	annualFee := models.AnnualFee{
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		Year:           body.Year,
		MemberCardUid:  body.MemberCardUid,
		ExpirationDate: body.ExpirationDate,
		// BillNumber:        body.BillNumber,
		Note:              body.Note,
		AnnualQuotaAmount: body.AnnualQuotaAmount,
		PrePaid:           body.PrePaid,
		PaidForfeit:       body.PaidForfeit,
		LastYearDebit:     body.LastYearDebit,
		TotalPaid:         body.TotalPaid,
		DaysPaid:          body.DaysPaid,
	}

	errC := annualFee.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, annualFee)
}

func (_ *CAnnualFee) GetListAnnualFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListAnnualFeeForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}
	if form.Year == 0 && form.MemberCardUid == "" {
		currentYearStr, errParseTime := utils.GetDateFromTimestampWithFormat(time.Now().Unix(), constants.YEAR_FORMAT)
		if errParseTime == nil {
			currentYearInt, errPInt := strconv.Atoi(currentYearStr)
			if errPInt == nil {
				form.Year = currentYearInt
			}
		}
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	annualFeeR := models.AnnualFee{
		PartnerUid:    form.PartnerUid,
		CourseUid:     form.CourseUid,
		MemberCardUid: form.MemberCardUid,
		Year:          form.Year,
	}
	list, totalData, total, err := annualFeeR.FindList(db, page, form.CardId)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total":    total,
		"data":     list,
		"sum_data": totalData,
	}

	okResponse(c, res)
}

func (_ *CAnnualFee) GetListAnnualFeeWithGroupMemberCard(c *gin.Context, prof models.CmsUser) {
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

	annualFeeR := models.AnnualFee{
		PartnerUid:    form.PartnerUid,
		CourseUid:     form.CourseUid,
		MemberCardUid: form.MemberCardUid,
		Year:          form.Year,
	}
	list, total, err := annualFeeR.FindListWithGroupMemberCard(db, page)
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

func (_ *CAnnualFee) UpdateAnnualFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	annualFeeIdStr := c.Param("id")
	annualFeeId, err := strconv.ParseInt(annualFeeIdStr, 10, 64)
	if err != nil || annualFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	annualFee := models.AnnualFee{}
	annualFee.Id = annualFeeId
	annualFee.PartnerUid = prof.PartnerUid
	annualFee.CourseUid = prof.CourseUid
	errF := annualFee.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.AnnualFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	//Check duplicated
	if body.Year != annualFee.Year && body.IsDuplicated(db) {
		response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	annualFee.AnnualQuotaAmount = body.AnnualQuotaAmount
	annualFee.PaidForfeit = body.PaidForfeit
	annualFee.LastYearDebit = body.LastYearDebit
	annualFee.TotalPaid = body.TotalPaid
	annualFee.DaysPaid = body.DaysPaid
	annualFee.PrePaid = body.PrePaid
	annualFee.PaidReduce = body.PaidReduce
	annualFee.ExpirationDate = body.ExpirationDate
	if body.Year > 0 {
		annualFee.Year = body.Year
	}

	errUdp := annualFee.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, annualFee)
}

func (_ *CAnnualFee) DeleteAnnualFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	annualFeeIdStr := c.Param("id")
	annualFeeId, err := strconv.ParseInt(annualFeeIdStr, 10, 64)
	if err != nil || annualFeeId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	annualFee := models.AnnualFee{}
	annualFee.Id = annualFeeId
	annualFee.PartnerUid = prof.PartnerUid
	annualFee.CourseUid = prof.CourseUid
	errF := annualFee.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := annualFee.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
