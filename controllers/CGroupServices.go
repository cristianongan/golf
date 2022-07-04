package controllers

import (
	"errors"
	"start/controllers/request"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"

	"github.com/gin-gonic/gin"
)

type CGroupServices struct{}

func (_ *CGroupServices) CreateGroupServices(c *gin.Context, prof models.CmsUser) {
	body := request.CreateGroupServicesBody{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	groupServices := model_service.GroupServices{}
	groupServices.GroupCode = body.GroupCode
	//Check Exits
	errFind := groupServices.FindFirst()
	if errFind == nil {
		response_message.DuplicateRecord(c, errors.New("Duplicate uid").Error())
		return
	}
	groupServices.GroupName = body.GroupName
	groupServices.Type = body.Type

	errC := groupServices.Create()

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, groupServices)
}

func (_ *CGroupServices) GetGroupServicesList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListGroupServicesForm{}
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

	groupServices := model_service.GroupServices{}
	if form.GroupCode != nil {
		groupServices.GroupCode = *form.GroupCode
	} else {
		groupServices.GroupCode = ""
	}

	if form.GroupName != nil {
		groupServices.GroupName = *form.GroupName
	} else {
		groupServices.GroupName = ""
	}

	if form.Type != nil {
		groupServices.Type = *form.Type
	} else {
		groupServices.Type = ""
	}

	list, total, err := groupServices.FindList(page)
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
