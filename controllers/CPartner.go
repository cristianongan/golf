package controllers

import (
	"errors"
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
	}
	partner.Uid = udpPartnerUid(body.Uid)
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
	partnerUidStr := c.Param("uid")
	// partnerUid, err := strconv.ParseInt(partnerUidStr, 10, 64) // Nếu uid là int64 mới cần convert
	if partnerUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	partner := models.Partner{}
	partner.Uid = partnerUidStr
	errF := partner.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := request.UpdatePartnerBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		partner.Name = body.Name
	}
	if body.Status != "" {
		partner.Status = body.Status
	}

	errUdp := partner.Update()
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, partner)
}

func (_ *CPartner) DeletePartner(c *gin.Context, prof models.CmsUser) {
	partnerUidStr := c.Param("uid")
	if partnerUidStr == "" {
		response_message.BadRequest(c, errors.New("uid not valid").Error())
		return
	}

	partner := models.Partner{}
	partner.Uid = partnerUidStr
	errF := partner.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := partner.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
