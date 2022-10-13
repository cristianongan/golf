package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CFbPromotionSet struct{}

func (_ *CFbPromotionSet) CreateFoodBeveragePromotionSet(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.FbPromotionSetBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
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
	errCourseFind := courseRequest.FindFirst(db)
	if errCourseFind != nil {
		response_message.BadRequest(c, errCourseFind.Error())
		return
	}

	fbList := []model_service.FBItem{}
	for _, item := range body.FBList {
		foodBeverage := model_service.FoodBeverage{
			FBCode:     item.Code,
			PartnerUid: body.PartnerUid,
			CourseUid:  body.CourseUid,
		}

		if err := foodBeverage.FindFirst(db); err != nil {
			response_message.BadRequest(c, errors.New(item.Code+" không tìm thấy ").Error())
			return
		}

		item := model_service.FBItem{
			FBCode:      foodBeverage.FBCode,
			Type:        foodBeverage.Type,
			EnglishName: foodBeverage.EnglishName,
			VieName:     foodBeverage.VieName,
			Price:       foodBeverage.Price,
			Unit:        foodBeverage.Unit,
			GroupCode:   foodBeverage.GroupCode,
			GroupName:   item.GroupName,
			Quantity:    item.Quantity,
		}

		fbList = append(fbList, item)
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}

	promotionSet := model_service.FbPromotionSet{
		ModelId:     base,
		CourseUid:   body.CourseUid,
		PartnerUid:  body.PartnerUid,
		VieName:     body.VieName,
		EnglishName: body.EnglishName,
		Discount:    body.Discount,
		Note:        body.Note,
		FBList:      fbList,
		Code:        body.Code,
		InputUser:   body.InputUser,
		Price:       body.Price,
	}

	promotionSet.Status = body.Status

	err := promotionSet.Create(db)
	if err != nil {
		log.Print("Create caddieNote error")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, promotionSet)
}

func (_ *CFbPromotionSet) GetListFoodBeveragepRomotionSet(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	promotionSetR := model_service.FbPromotionSetRequest{}
	promotionSetR.PartnerUid = form.PartnerUid
	promotionSetR.CourseUid = form.CourseUid
	promotionSetR.CodeOrName = form.CodeOrName
	promotionSetR.Status = form.Status

	list, total, err := promotionSetR.FindList(db, page)
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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	errF := promotionSetR.FindFirst(db)
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}
	if body.EnglishName != "" {
		promotionSetR.EnglishName = body.EnglishName
	}
	if body.VieName != "" {
		promotionSetR.VieName = body.VieName
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
	if body.Price > 0 {
		promotionSetR.Price = body.Price
	}

	var price float64 = 0
	if body.FBList != nil {
		fbList := model_service.FBSet{}
		for _, code := range body.FBList {
			foodBeverage := model_service.FoodBeverage{
				FBCode:     code,
				PartnerUid: promotionSetR.PartnerUid,
				CourseUid:  promotionSetR.CourseUid,
			}

			if err := foodBeverage.FindFirst(db); err != nil {
				response_message.BadRequest(c, errors.New(code+" không tìm thấy ").Error())
				return
			}

			item := model_service.FBItem{
				FBCode:      foodBeverage.FBCode,
				EnglishName: foodBeverage.EnglishName,
				VieName:     foodBeverage.VieName,
				Price:       foodBeverage.Price,
				Unit:        foodBeverage.Unit,
			}

			fbList = append(fbList, item)
			price += foodBeverage.Price
		}
		promotionSetR.FBList = fbList
	}

	err := promotionSetR.Update(db)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
func (_ *CFbPromotionSet) DeleteFoodBeveragePromotionSet(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	fbIdStr := c.Param("id")
	fbId, errId := strconv.ParseInt(fbIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	fbModel := model_service.FbPromotionSet{}
	fbModel.Id = fbId
	errF := fbModel.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := fbModel.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
