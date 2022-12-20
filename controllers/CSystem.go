package controllers

import (
	"errors"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CSystem struct{}

// Currency Paid
func (_ *CSystem) GetListCurencyRate(c *gin.Context, prof models.CmsUser) {
	currencyPaidGet := models.CurrencyPaid{}

	list, err := currencyPaidGet.FindAll()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": len(list),
		"data":  list,
	}

	okResponse(c, res)
}

// Nationality
func (_ *CSystem) GetListNationality(c *gin.Context, prof models.CmsUser) {
	nationalityGet := models.Nationality{}

	list, err := nationalityGet.FindAll()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": len(list),
		"data":  list,
	}

	okResponse(c, res)
}

// Category Type
func (_ *CSystem) GetListCategoryType(c *gin.Context, prof models.CmsUser) {
	cusTypesGet := models.CustomerType{}

	list, err := cusTypesGet.FindAll()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := map[string]interface{}{
		"total": len(list),
		"data":  list,
	}

	okResponse(c, res)
}

// ---------- Job -----------
func (_ *CSystem) CreateJob(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.SystemConfigJob{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	job := models.SystemConfigJob{}
	job.PartnerUid = body.PartnerUid
	job.CourseUid = body.CourseUid
	job.Name = body.Name

	//Check Exits
	errFind := job.FindFirst(db)
	if errFind == nil || job.Id > 0 {
		response_message.DuplicateRecord(c, errors.New("Duplicate name").Error())
		return
	}

	job.Status = body.Status

	errC := job.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, job)
}

func (_ *CSystem) GetListJob(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GeneralPageRequest{}
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

	jobR := models.SystemConfigJob{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := jobR.FindList(db, page)
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

func (_ *CSystem) UpdateJob(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	jobIdStr := c.Param("id")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || jobId == 0 {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	job := models.SystemConfigJob{}
	job.Id = jobId
	job.PartnerUid = prof.PartnerUid
	job.CourseUid = prof.CourseUid
	errF := job.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UpdatePartnerBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		job.Name = body.Name
	}
	if body.Status != "" {
		job.Status = body.Status
	}

	errUdp := job.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, job)
}

func (_ *CSystem) DeleteJob(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	jobIdStr := c.Param("id")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || jobId == 0 {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	job := models.SystemConfigJob{}
	job.Id = jobId
	job.PartnerUid = prof.PartnerUid
	job.CourseUid = prof.CourseUid
	errF := job.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := job.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

// ---------- Position -----------
func (_ *CSystem) CreatePosition(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.SystemConfigPosition{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	position := models.SystemConfigPosition{}
	position.Name = body.Name
	position.PartnerUid = body.PartnerUid
	position.CourseUid = body.CourseUid

	//Check Exits
	errFind := position.FindFirst(db)
	if errFind == nil || position.Id > 0 {
		response_message.DuplicateRecord(c, errors.New("Duplicate name").Error())
		return
	}

	position.Status = body.Status

	errC := position.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, position)
}

func (_ *CSystem) GetListPosition(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GeneralPageRequest{}
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

	positionR := models.SystemConfigPosition{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := positionR.FindList(db, page)
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

func (_ *CSystem) UpdatePosition(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	positionIdStr := c.Param("id")
	positionId, err := strconv.ParseInt(positionIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || positionId == 0 {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	position := models.SystemConfigPosition{}
	position.Id = positionId
	position.PartnerUid = prof.PartnerUid
	position.CourseUid = prof.CourseUid
	errF := position.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.SystemConfigPosition{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		position.Name = body.Name
	}
	if body.Status != "" {
		position.Status = body.Status
	}

	errUdp := position.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, position)
}

func (_ *CSystem) DeletePosition(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	positionIdStr := c.Param("id")
	positionId, err := strconv.ParseInt(positionIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || positionId == 0 {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	position := models.SystemConfigPosition{}
	position.Id = positionId
	position.PartnerUid = prof.PartnerUid
	position.CourseUid = prof.CourseUid
	errF := position.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := position.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

// ---------- Company Type -----------
func (_ *CSystem) CreateCompanyType(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.CompanyType{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	companyType := models.CompanyType{}
	companyType.Name = body.Name
	companyType.PartnerUid = body.PartnerUid
	companyType.CourseUid = body.CourseUid

	//Check Exits
	errFind := companyType.FindFirst(db)
	if errFind == nil || companyType.Id > 0 {
		response_message.DuplicateRecord(c, errors.New("Duplicate name").Error())
		return
	}

	companyType.Status = body.Status

	errC := companyType.Create(db)
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, companyType)
}

func (_ *CSystem) GetListCompanyType(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GeneralPageRequest{}
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

	companyTypeR := models.CompanyType{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := companyTypeR.FindList(db, page)
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

func (_ *CSystem) UpdateCompanyType(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	companyTypeIdStr := c.Param("id")
	companyTypeId, err := strconv.ParseInt(companyTypeIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || companyTypeId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	companyType := models.CompanyType{}
	companyType.Id = companyTypeId
	companyType.PartnerUid = prof.PartnerUid
	companyType.CourseUid = prof.CourseUid
	errF := companyType.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.CompanyType{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		companyType.Name = body.Name
	}
	if body.Status != "" {
		companyType.Status = body.Status
	}

	errUdp := companyType.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, companyType)
}

func (_ *CSystem) DeleteCompanyType(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	companyTypeIdStr := c.Param("id")
	companyTypeId, err := strconv.ParseInt(companyTypeIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || companyTypeId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	companyType := models.CompanyType{}
	companyType.Id = companyTypeId
	companyType.PartnerUid = prof.PartnerUid
	companyType.CourseUid = prof.CourseUid
	errF := companyType.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := companyType.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
