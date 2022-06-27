package controllers

import (
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBuggy struct{}

func (_ *CBuggy) CreateBuggy(c *gin.Context, prof models.CmsUser) {
	var body request.CreateBuggyBody
	if bindErr := c.BindJSON(&body); bindErr != nil {
		response_message.BadRequest(c, "")
		return
	}

	courseRequest := models.Course{}
	courseRequest.Uid = body.CourseUid
	errFind := courseRequest.FindFirst()
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	partnerRequest := models.Partner{}
	partnerRequest.Uid = body.PartnerUid
	partnerErrFind := partnerRequest.FindFirst()
	if partnerErrFind != nil {
		response_message.BadRequest(c, partnerErrFind.Error())
		return
	}

	buggyRequest := models.Buggy{}
	buggyRequest.CourseUid = body.CourseUid
	buggyRequest.PartnerUid = body.PartnerUid
	buggyRequest.Code = body.Code // Id Buggy vận hành
	errExist := buggyRequest.FindFirst()

	if errExist == nil && buggyRequest.ModelId.Id > 0 {
		response_message.BadRequest(c, "Code existed in course")
		return
	}

	base := models.ModelId{
		Status: constants.STATUS_ENABLE,
	}
	buggy := models.Buggy{
		ModelId:         base,
		Code:            body.Code,
		CourseUid:       body.CourseUid,
		PartnerUid:      body.PartnerUid,
		Origin:          body.Origin,
		Note:            body.Note,
		BuggyForVip:     body.BuggyForVip,
		MaintenanceFrom: body.MaintenanceFrom,
		MaintenanceTo:   body.MaintenanceTo,
		BuggyStatus:     body.BuggyStatus,
	}

	err := buggy.Create()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, buggy)
}

func (_ *CBuggy) GetBuggyList(c *gin.Context, prof models.CmsUser) {
	form := request.GetListBuggyForm{}
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

	buggyRequest := models.Buggy{}

	if form.Code != nil {
		buggyRequest.Code = *form.Code
	} else {
		buggyRequest.Code = ""
	}

	if form.BuggyStatus != nil {
		buggyRequest.BuggyStatus = *form.BuggyStatus
	} else {
		buggyRequest.BuggyStatus = ""
	}

	if form.BuggyForVip != nil {
		buggyRequest.BuggyForVip = *form.BuggyForVip
	}

	list, total, err := buggyRequest.FindList(page)

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

func (_ *CBuggy) DeleteBuggy(c *gin.Context, prof models.CmsUser) {
	buggyIdStr := c.Param("id")
	buggyId, errId := strconv.ParseInt(buggyIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	buggyRequest := models.Buggy{}
	buggyRequest.Id = buggyId
	errF := buggyRequest.FindFirst()

	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	err := buggyRequest.Delete()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CBuggy) UpdateBuggy(c *gin.Context, prof models.CmsUser) {
	buggyIdStr := c.Param("id")
	buggyId, errId := strconv.ParseInt(buggyIdStr, 10, 64)
	if errId != nil {
		response_message.BadRequest(c, errId.Error())
		return
	}

	var body request.UpdateBuggyBody
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	buggyRequest := models.Buggy{}
	buggyRequest.Id = buggyId

	errF := buggyRequest.FindFirst()
	if errF != nil {
		response_message.BadRequest(c, errF.Error())
		return
	}

	if body.Origin != nil {
		buggyRequest.Origin = *body.Origin
	}
	if body.Note != nil {
		buggyRequest.Note = *body.Note
	}

	err := buggyRequest.Update()
	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
