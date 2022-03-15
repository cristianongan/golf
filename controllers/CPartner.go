package controllers

import (
	"start/controllers/request"
	"start/models"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CPartner struct{}

func (_ *CPartner) CreatePartner(c *gin.Context, prof models.CmsUser) {
	body := request.CreatePartnerBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	partner := models.Partner{
		Name: body.Name,
		Code: body.Code,
	}
	partner.Status = body.Status

	errC := partner.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, partner)
}

func (_ *CPartner) GetListPartner(c *gin.Context, prof models.CmsUser) {
	form := request.GetListPartnerForm{}
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

	partnerR := models.Partner{}
	list, total, err := partnerR.FindList(page)
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

func (_ *CPartner) UpdatePartner(c *gin.Context, prof models.CmsUser) {
	body := request.UpdatePartnerBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

}
