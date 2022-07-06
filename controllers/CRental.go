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
	rentalRequest.SystemCode = body.SystemCode
	errExist := rentalRequest.FindFirst()

	if errExist == nil {
		response_message.BadRequest(c, "Rental Id existed in course")
		return
	}

	name := "" // tên default của proshop

	if body.EnglishName != "" {
		name = body.EnglishName
	} else {
		name = body.VieName
	}

	rental := model_service.Rental{
		RentalId:    body.RentalId,
		SystemCode:  body.SystemCode,
		PartnerUid:  body.PartnerUid,
		CourseUid:   body.CourseUid,
		EnglishName: body.EnglishName,
		RenPos:      body.RenPos,
		VieName:     body.VieName,
		GroupCode:   body.GroupCode,
		Unit:        body.Unit,
		Price:       body.Price,
		ByHoles:     body.ByHoles,
		ForPos:      body.ForPos,
		OnlyForRen:  body.OnlyForRen,
		InputUser:   body.InputUser,
		Name:        name,
	}
	rental.Status = body.Status

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

	rental := model_service.Rental{}
	rental.Id = rentalId
	errF := rental.FindFirst()
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
		rental.EnglishName = *body.EnglishName
	}
	if body.VieName != nil {
		rental.VieName = *body.VieName
	}
	if body.GroupCode != nil {
		rental.GroupCode = *body.GroupCode
	}
	if body.SystemCode != nil {
		rental.SystemCode = *body.SystemCode
	}
	if body.Unit != nil {
		rental.Unit = *body.Unit
	}
	if body.RenPos != nil {
		rental.RenPos = *body.RenPos
	}
	if body.Price != nil {
		rental.Price = *body.Price
	}
	if body.ByHoles != nil {
		rental.ByHoles = *body.ByHoles
	}
	if body.ForPos != nil {
		rental.ForPos = *body.ForPos
	}
	if body.OnlyForRen != nil {
		rental.OnlyForRen = *body.OnlyForRen
	}
	if body.Status != nil {
		rental.Status = *body.Status
	}
	if body.InputUser != nil {
		rental.InputUser = *body.InputUser
	}

	errUdp := rental.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, rental)
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
