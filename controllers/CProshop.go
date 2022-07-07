package controllers

import (
	"start/controllers/request"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CProshop struct{}

func (_ *CProshop) CreateProshop(c *gin.Context, prof models.CmsUser) {
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

	servicesRequest := model_service.GroupServices{}
	servicesRequest.GroupCode = body.GroupCode
	servicesErrFind := servicesRequest.FindFirst()
	if servicesErrFind != nil {
		response_message.BadRequest(c, servicesErrFind.Error())
		return
	}

	partnerRequest := models.Partner{}
	partnerRequest.Uid = body.PartnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		response_message.BadRequest(c, partnerErrFind.Error())
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseUid
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	ProshopRequest := model_service.Proshop{}
	ProshopRequest.CourseUid = body.CourseUid
	ProshopRequest.PartnerUid = body.PartnerUid
	ProshopRequest.ProShopId = body.ProshopId
	errExist := ProshopRequest.FindFirst()

	if errExist == nil {
		response_message.BadRequest(c, "F&B Id existed")
		return
	}

	name := "" // tên default của proshop

	if body.EnglishName != "" {
		name = body.EnglishName
	} else {
		name = body.VieName
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
		Name:          name,
		ProPrice:      body.ProPrice,
		PeopleDeposit: body.PeopleDeposit,
	}

	err := service.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, service)
}

func (_ *CProshop) GetListProshop(c *gin.Context, prof models.CmsUser) {
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
	if form.PartnerUid != nil {
		ProshopR.PartnerUid = *form.PartnerUid
	} else {
		ProshopR.PartnerUid = ""
	}
	if form.CourseUid != nil {
		ProshopR.CourseUid = *form.CourseUid
	} else {
		ProshopR.CourseUid = ""
	}
	if form.EnglishName != nil {
		ProshopR.EnglishName = *form.EnglishName
	} else {
		ProshopR.EnglishName = ""
	}
	if form.VieName != nil {
		ProshopR.VieName = *form.VieName
	} else {
		ProshopR.VieName = ""
	}
	if form.GroupCode != nil {
		ProshopR.GroupCode = *form.GroupCode
	} else {
		ProshopR.GroupCode = ""
	}
	if form.GroupName != nil {
		ProshopR.GroupName = *form.GroupCode
	} else {
		ProshopR.GroupName = ""
	}

	list, total, err := ProshopR.FindList(page)
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
	ProshopIdStr := c.Param("id")
	ProshopId, errId := strconv.ParseInt(ProshopIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	proshop := model_service.Proshop{}
	proshop.Id = ProshopId
	errF := proshop.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UpdateProshopBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.GroupCode != nil {
		proshop.GroupCode = *body.GroupCode
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
		proshop.AccountCode = *body.Note
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

	errUdp := proshop.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, proshop)
}

func (_ *CProshop) DeleteProshop(c *gin.Context, prof models.CmsUser) {
	fbIdStr := c.Param("id")
	fbId, errId := strconv.ParseInt(fbIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	fbModel := model_service.Proshop{}
	fbModel.Id = fbId
	errF := fbModel.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := fbModel.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
