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
	groupServices.SubType = body.SubType
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
	groupServices.CourseUid = form.CourseUid
	groupServices.PartnerUid = form.PartnerUid
	groupServices.Type = form.Type
	groupServices.GroupName = form.GroupName
	groupServices.GroupCode = form.GroupCode
	groupServices.SubType = form.SubType

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

func (_ *CGroupServices) GetGSAdvancedList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetGSAdvancedListForm{}
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
	groupServices.CourseUid = form.CourseUid
	groupServices.PartnerUid = form.PartnerUid
	groupServices.GroupName = form.GroupName

	list, total, err := groupServices.FindAdvancedList(db, page)
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	listData := make([]map[string]interface{}, len(list))

	for i, item := range list {
		if item.SubType == "" {
			//find all item in bill
			gsItem := model_service.GroupServices{}
			gsItem.CourseUid = item.CourseUid
			gsItem.PartnerUid = item.PartnerUid
			gsItem.SubType = item.GroupCode
			gsItem.GroupName = form.GroupName

			listGSItem, err := gsItem.FindAll(db)
			if err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}

			// Add infor to response
			listData[i] = map[string]interface{}{
				"infor":     item,
				"list_item": listGSItem,
			}
		} else {
			// find infor group cha
			gsInfor := model_service.GroupServices{}
			gsInfor.CourseUid = item.CourseUid
			gsInfor.PartnerUid = item.PartnerUid
			gsInfor.GroupCode = item.SubType

			err := gsInfor.FindFirst(db)
			if err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}

			//find all item con của group cha
			gsItem := model_service.GroupServices{}
			gsItem.CourseUid = item.CourseUid
			gsItem.PartnerUid = item.PartnerUid
			gsItem.SubType = item.SubType
			gsItem.GroupName = form.GroupName

			listGSItem, err := gsItem.FindAll(db)
			if err != nil {
				response_message.BadRequest(c, err.Error())
				return
			}

			// Add infor to response
			listData[i] = map[string]interface{}{
				"infor":     gsInfor,
				"list_item": listGSItem,
			}
		}
	}

	res := map[string]interface{}{
		"total": total,
		"data":  listData,
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

	if body.SubType != "" {
		groupServices.SubType = body.SubType
	}

	errUdp := groupServices.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, groupServices)
}
