package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	model_service "start/models/service"

	"github.com/gin-gonic/gin"
)

type CGolfService struct{}

func (_ *CGolfService) GetGolfServiceForReception(c *gin.Context, prof models.CmsUser) {
	form := request.GetGolfServiceForReceptionForm{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if form.PartnerUid == "" || form.CourseUid == "" || form.Type == "" {
		response_message.BadRequest(c, "data invalid")
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	if form.Type == constants.GOLF_SERVICE_RENTAL {
		// Get in rental
		rentalR := model_service.Rental{
			PartnerUid:  form.PartnerUid,
			CourseUid:   form.CourseUid,
			Type:        form.Type,
			Code:        form.Code,
			EnglishName: form.Name,
		}

		list, total, errRentalR := rentalR.FindList(page)

		if errRentalR != nil {
			response_message.InternalServerError(c, errRentalR.Error())
			return
		}

		res := map[string]interface{}{
			"total": total,
			"data":  list,
		}

		okResponse(c, res)
		return
	} else if form.Type == constants.GOLF_SERVICE_PROSHOP {
		// Get in proshop
		proshopR := model_service.Proshop{
			PartnerUid: form.PartnerUid,
			CourseUid:  form.CourseUid,
			Type:       form.Type,
			Code:       form.Code,
			Name:       form.Name,
		}

		list, total, errProshopR := proshopR.FindList(page)

		if errProshopR != nil {
			response_message.InternalServerError(c, errProshopR.Error())
			return
		}

		res := map[string]interface{}{
			"total": total,
			"data":  list,
		}

		okResponse(c, res)
		return
	} else if form.Type == constants.GOLF_SERVICE_RESTAURANT {
		// Get in restaurent
		restaurentR := model_service.Restaurent{
			PartnerUid: form.PartnerUid,
			CourseUid:  form.CourseUid,
			Type:       form.Type,
			Code:       form.Code,
			Name:       form.Name,
		}

		list, total, errRestaurentR := restaurentR.FindList(page)

		if errRestaurentR != nil {
			response_message.InternalServerError(c, errRestaurentR.Error())
			return
		}

		res := map[string]interface{}{
			"total": total,
			"data":  list,
		}

		okResponse(c, res)
		return
	} else if form.Type == constants.GOLF_SERVICE_KIOSK {
		// Get in kiosk
		kioskR := model_service.Kiosk{
			PartnerUid: form.PartnerUid,
			CourseUid:  form.CourseUid,
			Type:       form.Type,
			Code:       form.Code,
			Name:       form.Name,
		}

		list, total, errKioskR := kioskR.FindList(page)

		if errKioskR != nil {
			response_message.InternalServerError(c, errKioskR.Error())
			return
		}

		res := map[string]interface{}{
			"total": total,
			"data":  list,
		}

		okResponse(c, res)
		return
	}

	response_message.BadRequest(c, "type invalid")
	return
}
