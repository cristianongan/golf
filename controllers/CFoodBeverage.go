package controllers

import (
	"start/controllers/request"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

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
		Name:          name,
	}

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

	rentalR := model_service.FoodBeverage{}
	if form.PartnerUid != nil {
		rentalR.PartnerUid = *form.PartnerUid
	} else {
		rentalR.PartnerUid = ""
	}
	if form.CourseUid != nil {
		rentalR.CourseUid = *form.CourseUid
	} else {
		rentalR.CourseUid = ""
	}
	if form.EnglishName != nil {
		rentalR.EnglishName = *form.EnglishName
	} else {
		rentalR.EnglishName = ""
	}
	if form.VieName != nil {
		rentalR.VieName = *form.VieName
	} else {
		rentalR.VieName = ""
	}
	if form.GroupCode != nil {
		rentalR.GroupCode = *form.GroupCode
	} else {
		rentalR.GroupCode = ""
	}
	if form.Status != nil {
		rentalR.Status = *form.Status
	} else {
		rentalR.Status = ""
	}

	list, total, err := rentalR.FindList(page)
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

	foodBeverage.GroupCode = body.GroupCode
	foodBeverage.EnglishName = body.EnglishName
	foodBeverage.VieName = body.VieName
	foodBeverage.Unit = body.Unit
	foodBeverage.Price = body.Price
	foodBeverage.NetCost = body.NetCost
	foodBeverage.CostPrice = body.CostPrice
	foodBeverage.InternalPrice = body.InternalPrice
	foodBeverage.BarBeerPrice = body.BarBeerPrice
	foodBeverage.Barcode = body.Barcode
	foodBeverage.AccountCode = body.AccountCode
	foodBeverage.AloneKiosk = body.AloneKiosk
	foodBeverage.Note = body.Note
	foodBeverage.Status = body.Status
	foodBeverage.ForKiosk = body.ForKiosk
	foodBeverage.OpenFB = body.OpenFB
	foodBeverage.InMenuSet = body.InMenuSet
	foodBeverage.IsInventory = body.IsInventory

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
