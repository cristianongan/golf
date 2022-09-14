package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CFbPromotionSet struct{}

func (_ *CFbPromotionSet) CreateFoodBeveragePromotionSet(c *gin.Context, prof models.CmsUser) {
	body := request.FbPromotionSetBody{}
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

	var price float64 = 0

	for _, code := range body.FBList {
		foodBeverage := model_service.FoodBeverage{
			FBCode:     code,
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
		}

		if err := foodBeverage.FindFirst(); err != nil {
			response_message.BadRequest(c, errors.New(code+" không tìm thấy ").Error())
			return
		}
		price += foodBeverage.Price
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}

	promotionSet := model_service.FbPromotionSet{
		ModelId:    base,
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
		GroupCode:  body.GroupCode,
		SetName:    body.SetName,
		Discount:   body.Discount,
		Note:       body.Note,
		FBList:     body.FBList,
		Code:       body.Code,
		InputUser:  body.InputUser,
		Price:      price,
	}

	promotionSet.Status = body.Status

	err := promotionSet.Create()
	if err != nil {
		log.Print("Create caddieNote error")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, promotionSet)
}

func (_ *CFbPromotionSet) GetListFoodBeveragepRomotionSet(c *gin.Context, prof models.CmsUser) {
	form := request.GetListFbPromotionSetForm{}
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

	promotionSetR := model_service.FbPromotionSet{}
	if form.PartnerUid != nil {
		promotionSetR.PartnerUid = *form.PartnerUid
	} else {
		promotionSetR.PartnerUid = ""
	}
	if form.CourseUid != nil {
		promotionSetR.CourseUid = *form.CourseUid
	} else {
		promotionSetR.CourseUid = ""
	}
	if form.SetName != nil {
		promotionSetR.SetName = *form.SetName
	} else {
		promotionSetR.SetName = ""
	}
	if form.GroupCode != nil {
		promotionSetR.GroupCode = *form.GroupCode
	} else {
		promotionSetR.GroupCode = ""
	}
	if form.Status != nil {
		promotionSetR.Status = *form.Status
	} else {
		promotionSetR.Status = ""
	}

	list, total, err := promotionSetR.FindList(page)
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

func (_ *CFbPromotionSet) UpdatePromotionSet(c *gin.Context, prof models.CmsUser) {
	idStr := c.Param("id")
	Id, errId := strconv.ParseInt(idStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateFbPromotionSet
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	promotionSetR := model_service.FbPromotionSet{}
	promotionSetR.Id = Id

	errF := promotionSetR.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.SetName != nil {
		promotionSetR.SetName = *body.SetName
	}
	if body.Note != nil {
		promotionSetR.Note = *body.Note
	}
	if body.Discount != nil {
		promotionSetR.Discount = *body.Discount
	}
	if body.Status != nil {
		promotionSetR.Status = *body.Status
	}
	if body.FBList != nil {
		promotionSetR.FBList = *body.FBList
	}

	err := promotionSetR.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
func (_ *CFbPromotionSet) DeleteFoodBeveragePromotionSet(c *gin.Context, prof models.CmsUser) {
	fbIdStr := c.Param("id")
	fbId, errId := strconv.ParseInt(fbIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	fbModel := model_service.FbPromotionSet{}
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
