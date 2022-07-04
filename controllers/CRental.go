package controllers

import (
	"log"
	"start/controllers/request"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CRental struct{}

func (_ *CRental) CreateRental(c *gin.Context, prof models.CmsUser) {
	body := request.CreateRentalBody{}
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

	rentalRequest := model_service.Rental{}
	rentalRequest.CourseUid = body.CourseUid
	rentalRequest.PartnerUid = body.PartnerUid
	rentalRequest.Code = body.Code
	errExist := rentalRequest.FindFirst()

	if errExist == nil {
		response_message.BadRequest(c, "Rental Id existed in course")
		return
	}

	rental := model_service.Rental{
		Code:         body.Code,
		PartnerUid:   body.PartnerUid,
		CourseUid:    body.CourseUid,
		EnglishName:  body.EnglishName,
		RenPos:       body.RenPos,
		VieName:      body.VieName,
		GroupId:      body.GroupId,
		GroupCode:    body.GroupCode,
		GroupName:    body.GroupName,
		Unit:         body.Unit,
		Price:        body.Price,
		ByHoles:      body.ByHoles,
		ForPos:       body.ForPos,
		OnlyForRen:   body.OnlyForRen,
		RentalStatus: body.RentalStatus,
		InputUser:    body.InputUser,
	}

	err := rental.Create()
	if err != nil {
		log.Print("Caddie.Create()")
		response_message.InternalServerError(c, err.Error())
		return
	}

	okResponse(c, rental)
}

func (_ *CRental) GetListRental(c *gin.Context, prof models.CmsUser) {
	form := request.GetListRentalForm{}
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

	rentalR := model_service.Rental{}
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
	if form.RentalStatus != nil {
		rentalR.RentalStatus = *form.RentalStatus
	} else {
		rentalR.RentalStatus = ""
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

func (_ *CRental) UpdateRental(c *gin.Context, prof models.CmsUser) {
	rentalIdStr := c.Param("id")
	rentalId, errId := strconv.ParseInt(rentalIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	partner := model_service.Rental{}
	partner.Id = rentalId
	errF := partner.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UpdateRentalBody{}
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
	if body.RentalStatus != nil {
		partner.RentalStatus = *body.RentalStatus
	}
	if body.ByHoles != nil {
		partner.ByHoles = *body.ByHoles
	}
	if body.ForPos != nil {
		partner.ForPos = *body.ForPos
	}
	if body.EnglishName != nil {
		partner.OnlyForRen = *body.OnlyForRen
	}
	if body.Price != nil {
		partner.Price = *body.Price
	}

	errUdp := partner.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, partner)
}

func (_ *CRental) DeleteRental(c *gin.Context, prof models.CmsUser) {
	rentalIdStr := c.Param("id")
	rentalId, errId := strconv.ParseInt(rentalIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	rental := model_service.Rental{}
	rental.Id = rentalId
	errF := rental.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := rental.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}