package controllers

import (
	"errors"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	model_service "start/models/service"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CGroupServices struct{}

func (_ *CGroupServices) CreateGroupServices(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := request.CreateGroupServicesBody{}
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

	groupServices := model_service.GroupServices{}
	groupServices.GroupCode = body.GroupCode
	groupServices.CourseUid = body.CourseUid
	groupServices.PartnerUid = body.PartnerUid
	//Check Exits
	errFind := groupServices.FindFirst(db)
	if errFind == nil {
		response_message.DuplicateRecord(c, errors.New("Duplicate uid").Error())
		return
	}
	groupServices.GroupName = body.GroupName
	groupServices.Type = body.Type
	groupServices.DetailGroup = body.DetailGroup

	errC := groupServices.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, groupServices)
}

func (_ *CGroupServices) GetGroupServicesList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	if form.PartnerUid != nil {
		groupServices.PartnerUid = *form.PartnerUid
	} else {
		groupServices.PartnerUid = ""
	}

	if form.CourseUid != nil {
		groupServices.CourseUid = *form.CourseUid
	} else {
		groupServices.CourseUid = ""
	}

	list, total, err := groupServices.FindList(db, page)
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

func (_ *CGroupServices) DeleteServices(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	serviceIdP := c.Param("id")
	serviceId, errId := strconv.ParseInt(serviceIdP, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	groupServices := model_service.GroupServices{}
	groupServices.Id = serviceId
	errF := groupServices.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := groupServices.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

func (_ *CGroupServices) UpdateServices(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	serviceIdP := c.Param("id")
	serviceId, errId := strconv.ParseInt(serviceIdP, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	groupServices := model_service.GroupServices{}
	groupServices.Id = serviceId
	errF := groupServices.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := model_service.GroupServices{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.GroupCode != "" {
		groupServices.GroupCode = body.GroupCode
	}

	if body.GroupName != "" {
		groupServices.GroupName = body.GroupName
	}

	if body.DetailGroup != "" {
		groupServices.DetailGroup = body.DetailGroup
	}

	if body.Type != "" {
		groupServices.Type = body.Type
	}

	errUdp := groupServices.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, groupServices)
}
