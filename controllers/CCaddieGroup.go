package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CCaddieGroup struct {
}

func (_ CCaddieGroup) GetCaddieGroupList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	query := request.GetCaddieGroupList{}
	if err := c.ShouldBind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	caddieGroup := models.CaddieGroup{
		PartnerUid: query.PartnerUid,
		CourseUid:  query.CourseUid,
	}

	list, total, err := caddieGroup.FindList(db, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	res := response.PageResponse{
		Total: total,
		Data:  list,
	}

	c.JSON(200, res)
}

func (_ CCaddieGroup) CreateCaddieGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateCaddieGroupBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieGroup BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	//TODO: validate group_code

	caddieGroup := models.CaddieGroup{
		Code:       body.GroupCode,
		Name:       body.GroupName,
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
	}

	if err := caddieGroup.ValidateCreate(db); err == nil {
		response_message.BadRequest(c, "This caddie group is exist")
		return
	}

	if err := caddieGroup.Create(db); err != nil {
		log.Print("CreateCaddieGroup.Create()")
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Add log
	dateAction, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	opLog := models.OperationLog{
		PartnerUid:  prof.PartnerUid,
		CourseUid:   prof.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_CADDIE,
		Function:    constants.OP_LOG_FUNCTION_GROUP_MANAGEMENT,
		Action:      constants.OP_LOG_ACTION_CREATE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{},
		ValueNew:    models.JsonDataLog{Data: caddieGroup},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: dateAction,
	}
	go createOperationLog(opLog)

	c.JSON(200, caddieGroup)
}

func (_ CCaddieGroup) AddCaddieToGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.AddCaddieToGroupBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("AddCaddieToGroup BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	// validate caddie_group
	caddieGroup := models.CaddieGroup{
		Code:       body.GroupCode,
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
	}

	if err := caddieGroup.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieList := models.CaddieList{}
	caddieList.PartnerUid = prof.PartnerUid
	caddieList.CourseUid = prof.CourseUid
	caddieList.CaddieCodeList = body.CaddieList
	list, err := caddieList.FindListWithoutPage(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, item := range list {
		item.GroupId = caddieGroup.Id
		if err := item.Update(db); err != nil {
			continue
		}
	}

	okRes(c)
}

func (_ CCaddieGroup) DeleteCaddieGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddieGroupId, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	caddieList := models.CaddieList{}
	caddieList.PartnerUid = prof.PartnerUid
	caddieList.CourseUid = prof.CourseUid
	caddieList.GroupId = caddieGroupId
	list, err := caddieList.FindListWithoutPage(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	hasError := false

	for _, caddie := range list {
		caddie.GroupId = 0
		if err := caddie.Update(db); err != nil {
			response_message.BadRequest(c, err.Error())
			hasError = true
			break
		}
	}

	if hasError {
		return
	}

	caddieGroup := models.CaddieGroup{}
	caddieGroup.Id = caddieGroupId

	if err := caddieGroup.Delete(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	// Add log
	dateAction, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	opLog := models.OperationLog{
		PartnerUid:  prof.PartnerUid,
		CourseUid:   prof.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_CADDIE,
		Function:    constants.OP_LOG_FUNCTION_GROUP_MANAGEMENT,
		Action:      constants.OP_LOG_ACTION_DELETE,
		Body:        models.JsonDataLog{Data: caddieGroupId},
		ValueOld:    models.JsonDataLog{Data: caddieGroup},
		ValueNew:    models.JsonDataLog{},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: dateAction,
	}
	go createOperationLog(opLog)

	okRes(c)
}

func (_ CCaddieGroup) MoveCaddieToGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.MoveCaddieToGroupBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("MoveCaddieToGroup BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	// validate caddie_group
	caddieGroup := models.CaddieGroup{
		Code:       body.GroupCode,
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
	}

	if err := caddieGroup.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieList := models.CaddieList{}
	caddieList.PartnerUid = prof.PartnerUid
	caddieList.CourseUid = prof.CourseUid
	caddieList.CaddieCodeList = body.CaddieList
	list, err := caddieList.FindListWithoutPage(db)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	//data old
	listOld := list

	for _, item := range list {
		item.GroupId = caddieGroup.Id
		if err := item.Update(db); err != nil {
			continue
		}
	}

	// Add log
	dateAction, _ := utils.GetBookingDateFromTimestamp(utils.GetTimeNow().Unix())

	opLog := models.OperationLog{
		PartnerUid:  prof.PartnerUid,
		CourseUid:   prof.CourseUid,
		UserName:    prof.UserName,
		UserUid:     prof.Uid,
		Module:      constants.OP_LOG_MODULE_CADDIE,
		Function:    constants.OP_LOG_FUNCTION_GROUP_MANAGEMENT,
		Action:      constants.OP_LOG_ACTION_MOVE,
		Body:        models.JsonDataLog{Data: body},
		ValueOld:    models.JsonDataLog{Data: listOld},
		ValueNew:    models.JsonDataLog{Data: list},
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		BookingDate: dateAction,
	}
	go createOperationLog(opLog)

	okRes(c)
}

func (_ CCaddieGroup) UpdateGroupCaddies(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body []request.UpdateCaddieGroupBody
	if err := c.BindJSON(&body); err != nil {
		response_message.BadRequest(c, "")
		return
	}

	var res struct {
		IdSuccess []int64
		IdFailse  []int64
	}

	for _, v := range body {
		caddie := models.Caddie{}
		caddie.Id = v.Id
		errFind := caddie.FindFirst(db)
		if errFind == nil {
			caddie.GroupId = v.GroupId
			errUpdate := caddie.Update(db)
			if errUpdate != nil {
				res.IdFailse = append(res.IdFailse, v.Id)
			} else {
				res.IdSuccess = append(res.IdSuccess, v.Id)
			}
		} else {
			res.IdFailse = append(res.IdFailse, v.Id)
		}
	}

	okResponse(c, res)
}
