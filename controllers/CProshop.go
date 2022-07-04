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
	ProshopRequest.ProCode = body.ProCode
	errExist := ProshopRequest.FindFirst()

	if errExist == nil {
		response_message.BadRequest(c, "F&B Id existed")
		return
	}

	service := model_service.Proshop{
		PartnerUid:  body.PartnerUid,
		GroupId:     body.GroupId,
		CourseUid:   body.CourseUid,
		GroupName:   body.GroupName,
		ProCode:     body.ProCode,
		EnglishName: body.EnglishName,
		VieName:     body.VieName,
		Unit:        body.Unit,
		Price:       body.Price,
		NetCost:     body.NetCost,
		CostPrice:   body.CostPrice,
		Barcode:     body.Barcode,
		AccountCode: body.AccountCode,
		Note:        body.Note,
		ForKiosk:    body.ForKiosk,
		IsInventory: body.IsInventory,
		Brand:       body.Brand,
		UserUpdate:  body.UserUpdate,
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

	ProshopR := model_service.Proshop{}
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
	if form.GroupId != nil {
		ProshopR.GroupId = *form.GroupId
	} else {
		ProshopR.GroupId = ""
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

	partner := model_service.Proshop{}
	partner.Id = ProshopId
	errF := partner.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UpdateProshopBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.EnglishName != nil {
		partner.EnglishName = *body.EnglishName
	}
	if body.VieName != nil {
		partner.VieName = *body.VieName
	}
	// if body.ProshopStatus != nil {
	// 	partner.ProshopStatus = *body.ProshopStatus
	// }
	// if body.ByHoles != nil {
	// 	partner.ByHoles = *body.ByHoles
	// }
	// if body.ForPos != nil {
	// 	partner.ForPos = *body.ForPos
	// }
	// if body.EnglishName != nil {
	// 	partner.OnlyForRen = *body.OnlyForRen
	// }
	// if body.Price != nil {
	// 	partner.Price = *body.Price
	// }

	errUdp := partner.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, partner)
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
