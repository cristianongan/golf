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

type CCompany struct{}

func (_ *CCompany) CreateCompany(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.Company{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.Code == "" || body.Name == "" || body.CompanyTypeId <= 0 || body.PartnerUid == "" || body.CourseUid == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// Check company type
	companyType := models.CompanyType{}
	companyType.Id = body.CompanyTypeId
	errF := companyType.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	company := models.Company{}
	company.PartnerUid = body.PartnerUid
	company.CourseUid = body.CourseUid
	company.Code = body.Code

	// Check duplicate code trong 1 hãng
	if company.IsDuplicated(db) {
		response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	company.Name = body.Name
	company.Status = body.Status
	company.Address = body.Address
	company.Fax = body.Fax
	company.FaxCode = body.FaxCode
	company.CompanyTypeId = companyType.Id
	company.CompanyTypeName = companyType.Name

	errC := company.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, company)
}

func (_ *CCompany) GetListCompany(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListCompanyForm{}
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

	companyR := models.Company{
		PartnerUid:    form.PartnerUid,
		CourseUid:     form.CourseUid,
		Name:          form.Name,
		CompanyTypeId: form.CompanyTypeId,
		Phone:         form.Phone,
		Code:          form.Code,
	}
	companyR.Status = form.Status
	list, total, err := companyR.FindList(db, page)
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

func (_ *CCompany) UpdateCompany(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	companyIdStr := c.Param("id")
	companyId, err := strconv.ParseInt(companyIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || companyId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	company := models.Company{}
	company.Id = companyId
	company.PartnerUid = prof.PartnerUid
	company.CourseUid = prof.CourseUid
	errF := company.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.Company{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Code != "" && body.Code != company.Code {
		if body.IsDuplicated(db) {
			response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
		company.Code = body.Code
	}

	if body.Name != "" {
		company.Name = body.Name
	}
	if body.Status != "" {
		company.Status = body.Status
	}
	if body.Address != "" {
		company.Address = body.Address
	}
	if body.Phone != "" {
		company.Phone = body.Phone
	}
	if body.Fax != "" {
		company.Fax = body.Fax
	}
	if body.FaxCode != "" {
		company.FaxCode = body.FaxCode
	}

	errUdp := company.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, company)
}

func (_ *CCompany) DeleteCompany(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	companyIdStr := c.Param("id")
	companyId, err := strconv.ParseInt(companyIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || companyId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	company := models.Company{}
	company.Id = companyId
	company.PartnerUid = prof.PartnerUid
	company.CourseUid = prof.CourseUid
	errF := company.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := company.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
