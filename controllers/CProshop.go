package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_service "start/models/service"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CProshop struct{}

func (_ *CProshop) CreateProshop(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateProshopBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.CourseUid == "" {
		response_message.BadRequest(c, "Course Uid not empty")
		return
	}

	if body.PartnerUid == "" {
		response_message.BadRequest(c, "Partner Uid not empty")
		return
	}

	if body.GroupCode == "" {
		response_message.BadRequest(c, "Group Code not empty")
		return
	}

	servicesRequest := model_service.GroupServices{}
	servicesRequest.GroupCode = body.GroupCode
	servicesErrFind := servicesRequest.FindFirst(db)
	if servicesErrFind != nil {
		response_message.BadRequest(c, "GroupCode not existed")
		return
	}

	partnerRequest := models.Partner{}
	partnerRequest.Uid = body.PartnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		response_message.BadRequest(c, "Partner not existed")
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseUid
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, "Course not existed")
		return
	}

	ProshopRequest := model_service.Proshop{}
	ProshopRequest.CourseUid = body.CourseUid
	ProshopRequest.PartnerUid = body.PartnerUid
	ProshopRequest.ProShopId = body.ProshopId
	errExist := ProshopRequest.FindFirst(db)

	if errExist == nil {
		response_message.BadRequest(c, "Proshop Id existed")
		return
	}

	service := model_service.Proshop{
		ProShopId:     body.ProshopId,
		PartnerUid:    body.PartnerUid,
		CourseUid:     body.CourseUid,
		GroupCode:     body.GroupCode,
		EnglishName:   body.EnglishName,
		VieName:       body.VieName,
		Unit:          body.Unit,
		Price:         body.Price,
		NetCost:       body.NetCost,
		CostPrice:     body.CostPrice,
		Barcode:       body.Barcode,
		AccountCode:   body.AccountCode,
		Note:          body.Note,
		ForKiosk:      body.ForKiosk,
		IsInventory:   body.IsInventory,
		IsDeposit:     body.IsDeposit,
		Brand:         body.Brand,
		UserUpdate:    body.UserUpdate,
		Name:          body.VieName,
		ProPrice:      body.ProPrice,
		PeopleDeposit: body.PeopleDeposit,
		Type:          body.Type,
		TaxCode:       body.TaxCode,
	}

	err := service.Create(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_SYSTEM_GOLFFEE,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Function:    constants.OP_LOG_FUNCTION_PROSHOP_SYSTEM,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: service},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: utils.GetCurrentDay1(),
	}

	go createOperationLog(opLog)

	okResponse(c, service)
}

func (_ *CProshop) GetListProshop(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListProshopForm{}
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

	ProshopR := model_service.ProshopRequest{}
	ProshopR.PartnerUid = form.PartnerUid
	ProshopR.CourseUid = form.CourseUid
	ProshopR.EnglishName = form.EnglishName
	ProshopR.VieName = form.VieName
	ProshopR.GroupCode = form.GroupCode
	ProshopR.Type = form.Type
	ProshopR.CodeOrName = form.CodeOrName

	list, total, err := ProshopR.FindList(db, page)
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

func (_ *CProshop) UpdateProshop(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	ProshopIdStr := c.Param("id")
	ProshopId, errId := strconv.ParseInt(ProshopIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	proshop := model_service.Proshop{}
	proshop.Id = ProshopId
	proshop.PartnerUid = prof.PartnerUid
	proshop.CourseUid = prof.CourseUid
	errF := proshop.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	oldProshop := proshop

	body := request.UpdateProshopBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.EnglishName != nil {
		proshop.EnglishName = *body.EnglishName
	}
	if body.VieName != nil {
		proshop.VieName = *body.VieName
	}
	if body.Unit != nil {
		proshop.Unit = *body.Unit
	}
	if body.Price != nil {
		proshop.Price = *body.Price

		// Update lại giá trong Kho
		go func() {
			cKioskInventory := CKioskInventory{}
			cKioskInventory.UpdatePriceForItem(db, prof.PartnerUid, prof.CourseUid, proshop.ProShopId, proshop.Price)
		}()
	}
	if body.NetCost != nil {
		proshop.NetCost = *body.NetCost
	}
	if body.CostPrice != nil {
		proshop.CostPrice = *body.CostPrice
	}
	if body.Barcode != nil {
		proshop.Barcode = *body.Barcode
	}
	if body.Brand != nil {
		proshop.Brand = *body.Brand
	}
	if body.PeopleDeposit != nil {
		proshop.PeopleDeposit = *body.PeopleDeposit
	}
	if body.AccountCode != nil {
		proshop.AccountCode = *body.AccountCode
	}
	if body.Note != nil {
		proshop.Note = *body.Note
	}
	if body.ForKiosk != nil {
		proshop.ForKiosk = *body.ForKiosk
	}
	if body.ProPrice != nil {
		proshop.ProPrice = *body.ProPrice
	}
	if body.IsDeposit != nil {
		proshop.IsDeposit = *body.IsDeposit
	}
	if body.IsInventory != nil {
		proshop.IsInventory = *body.IsInventory
	}
	if body.UserUpdate != nil {
		proshop.UserUpdate = *body.UserUpdate
	}
	if body.Type != "" {
		proshop.Type = body.Type
	}
	if body.GroupName != "" {
		proshop.GroupName = body.GroupName
	}
	if body.GroupCode != "" {
		proshop.GroupCode = body.GroupCode
	}
	if body.TaxCode != "" {
		proshop.TaxCode = body.TaxCode
	}

	errUdp := proshop.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  proshop.PartnerUid,
		CourseUid:   proshop.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_SYSTEM_GOLFFEE,
		Action:      constants.OP_LOG_ACTION_UPDATE,
		Function:    constants.OP_LOG_FUNCTION_PROSHOP_SYSTEM,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: oldProshop},
		ValueNew:    models.JsonDataLog{Data: proshop},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: utils.GetCurrentDay1(),
	}

	go createOperationLog(opLog)

	okResponse(c, proshop)
}

func (_ *CProshop) DeleteProshop(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	fbIdStr := c.Param("id")
	fbId, errId := strconv.ParseInt(fbIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	proshop := model_service.Proshop{}
	proshop.Id = fbId
	proshop.PartnerUid = prof.PartnerUid
	proshop.CourseUid = prof.CourseUid
	errF := proshop.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := proshop.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	//Add log
	opLog := models.OperationLog{
		PartnerUid:  proshop.PartnerUid,
		CourseUid:   proshop.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_SYSTEM_GOLFFEE,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Function:    constants.OP_LOG_FUNCTION_PROSHOP_SYSTEM,
		Body:        models.JsonDataLog{Data: fbIdStr},
		ValueOld:    models.JsonDataLog{Data: proshop},
		ValueNew:    models.JsonDataLog{},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: utils.GetCurrentDay1(),
	}

	go createOperationLog(opLog)

	okRes(c)
}
