package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"
)

type CCaddieGroup struct {
}

func (_ CCaddieGroup) GetCaddieGroupList(c *gin.Context, prof models.CmsUser) {
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

	caddieGroup := models.CaddieGroup{}

	list, total, err := caddieGroup.FindList(page)

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
	var body request.CreateCaddieGroupBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieGroup BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	//TODO: validate group_code

	caddieGroup := models.CaddieGroup{
		Code:       body.GroupCode,
		PartnerUid: prof.PartnerUid,
		CourseUid:  prof.CourseUid,
	}

	if err := caddieGroup.FindFirst(); err == nil {
		response_message.BadRequest(c, "This caddie group is exist")
		return
	}

	caddieGroup.Name = body.GroupName

	if err := caddieGroup.Create(); err != nil {
		log.Print("CreateCaddieGroup.Create()")
		response_message.InternalServerError(c, err.Error())
		return
	}

	c.JSON(200, caddieGroup)
}

func (_ CCaddieGroup) AddCaddieToGroup(c *gin.Context, prof models.CmsUser) {
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

	if err := caddieGroup.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieList := models.CaddieList{}
	caddieList.PartnerUid = prof.PartnerUid
	caddieList.CourseUid = prof.CourseUid
	caddieList.CaddieCodeList = body.CaddieList
	list, err := caddieList.FindListWithoutPage()

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, item := range list {
		item.GroupId = caddieGroup.Id
		if err := item.Update(); err != nil {
			continue
		}
	}

	okRes(c)
}

func (_ CCaddieGroup) DeleteCaddieGroup(c *gin.Context, prof models.CmsUser) {
	caddieGroupId, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	caddieList := models.CaddieList{}
	caddieList.PartnerUid = prof.PartnerUid
	caddieList.CourseUid = prof.CourseUid
	caddieList.GroupId = caddieGroupId
	list, err := caddieList.FindListWithoutPage()

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	hasError := false

	for _, caddie := range list {
		caddie.GroupId = 0
		if err := caddie.Update(); err != nil {
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

	if err := caddieGroup.Delete(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ CCaddieGroup) MoveCaddieToGroup(c *gin.Context, prof models.CmsUser) {
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

	if err := caddieGroup.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieList := models.CaddieList{}
	caddieList.PartnerUid = prof.PartnerUid
	caddieList.CourseUid = prof.CourseUid
	caddieList.CaddieCodeList = body.CaddieList
	list, err := caddieList.FindListWithoutPage()

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, item := range list {
		item.GroupId = caddieGroup.Id
		if err := item.Update(); err != nil {
			continue
		}
	}

	okRes(c)
}
