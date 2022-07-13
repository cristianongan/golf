package controllers

import (
	"start/controllers/request"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CFoodBeverage struct{}

func (_ *CFoodBeverage) CreateFoodBeverage(c *gin.Context, prof models.CmsUser) {
	body := request.CreateFoodBeverageBody{}
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
	errCourseFind := courseRequest.FindFirst()
	if errCourseFind != nil {
		response_message.BadRequest(c, errCourseFind.Error())
		return
	}

	foodBeverageRequest := model_service.FoodBeverage{}
	foodBeverageRequest.CourseUid = body.CourseUid
	foodBeverageRequest.PartnerUid = body.PartnerUid
	foodBeverageRequest.FBCode = body.FBCode
	errExist := foodBeverageRequest.FindFirst()

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

	service := model_service.FoodBeverage{
		PartnerUid:    body.PartnerUid,
		CourseUid:     body.CourseUid,
		GroupCode:     body.GroupCode,
		FBCode:        body.FBCode,
		EnglishName:   body.EnglishName,
		VieName:       body.VieName,
		Unit:          body.Unit,
		Price:         body.Price,
		NetCost:       body.NetCost,
		CostPrice:     body.CostPrice,
		Barcode:       body.Barcode,
		AccountCode:   body.AccountCode,
		BarBeerPrice:  body.BarBeerPrice,
		Note:          body.Note,
		ForKiosk:      body.ForKiosk,
		OpenFB:        body.OpenFB,
		AloneKiosk:    body.AloneKiosk,
		InMenuSet:     body.InMenuSet,
		IsInventory:   body.IsInventory,
		InternalPrice: body.InternalPrice,
		IsKitchen:     body.IsKitchen,
		Name:          name,
	}
	service.Status = body.Status

	err := service.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, service)
}

func (_ *CFoodBeverage) GetListFoodBeverage(c *gin.Context, prof models.CmsUser) {
	form := request.GetListFoodBeverageForm{}
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

	fbR := model_service.FoodBeverageRequest{}
	if form.PartnerUid != nil {
		fbR.PartnerUid = *form.PartnerUid
	} else {
		fbR.PartnerUid = ""
	}
	if form.CourseUid != nil {
		fbR.CourseUid = *form.CourseUid
	} else {
		fbR.CourseUid = ""
	}
	if form.EnglishName != nil {
		fbR.EnglishName = *form.EnglishName
	} else {
		fbR.EnglishName = ""
	}
	if form.VieName != nil {
		fbR.VieName = *form.VieName
	} else {
		fbR.VieName = ""
	}
	if form.GroupCode != nil {
		fbR.GroupCode = *form.GroupCode
	} else {
		fbR.GroupCode = ""
	}
	if form.Status != nil {
		fbR.Status = *form.Status
	} else {
		fbR.Status = ""
	}
	if form.FBCodeList != nil {
		fbR.FBCodeList = strings.Split(*form.FBCodeList, ",")
	}

	list, total, err := fbR.FindList(page)
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

func (_ *CFoodBeverage) UpdateFoodBeverage(c *gin.Context, prof models.CmsUser) {
	rentalIdStr := c.Param("id")
	rentalId, errId := strconv.ParseInt(rentalIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	foodBeverage := model_service.FoodBeverage{}
	foodBeverage.Id = rentalId
	errF := foodBeverage.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UpdateFoodBeverageBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.GroupCode != nil {
		foodBeverage.GroupCode = *body.GroupCode
	}
	if body.EnglishName != nil {
		foodBeverage.EnglishName = *body.EnglishName
	}
	if body.VieName != nil {
		foodBeverage.VieName = *body.VieName
	}
	if body.Unit != nil {
		foodBeverage.Unit = *body.Unit
	}
	if body.Price != nil {
		foodBeverage.Price = *body.Price
	}
	if body.NetCost != nil {
		foodBeverage.NetCost = *body.NetCost
	}
	if body.CostPrice != nil {
		foodBeverage.CostPrice = *body.CostPrice
	}
	if body.InternalPrice != nil {
		foodBeverage.InternalPrice = *body.InternalPrice
	}
	if body.Barcode != nil {
		foodBeverage.Barcode = *body.Barcode
	}
	if body.AccountCode != nil {
		foodBeverage.AccountCode = *body.AccountCode
	}
	if body.AloneKiosk != nil {
		foodBeverage.AloneKiosk = *body.AloneKiosk
	}
	if body.Note != nil {
		foodBeverage.Note = *body.Note
	}
	if body.Status != nil {
		foodBeverage.Status = *body.Status
	}
	if body.ForKiosk != nil {
		foodBeverage.ForKiosk = *body.ForKiosk
	}
	if body.OpenFB != nil {
		foodBeverage.OpenFB = *body.OpenFB
	}
	if body.InMenuSet != nil {
		foodBeverage.InMenuSet = *body.InMenuSet
	}
	if body.IsInventory != nil {
		foodBeverage.IsInventory = *body.IsInventory
	}
	if body.IsKitchen != nil {
		foodBeverage.IsKitchen = *body.IsKitchen
	}
	if body.UserUpdate != nil {
		foodBeverage.UserUpdate = *body.UserUpdate
	}

	errUdp := foodBeverage.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, foodBeverage)
}

func (_ *CFoodBeverage) DeleteFoodBeverage(c *gin.Context, prof models.CmsUser) {
	fbIdStr := c.Param("id")
	fbId, errId := strconv.ParseInt(fbIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	fbModel := model_service.FoodBeverage{}
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
