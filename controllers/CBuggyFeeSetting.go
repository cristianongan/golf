package controllers

import (
	"errors"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CBuggyFeeSetting struct{}

func (_ *CBuggyFeeSetting) CreateBuggyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.BuggyFeeSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	buggyFeeSetting := models.BuggyFeeSetting{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		TypeName:   body.TypeName,
	}
	buggyFeeSetting.Status = body.Status

	errC := buggyFeeSetting.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, buggyFeeSetting)
}

func (_ *CBuggyFeeSetting) GetBuggyFeeSettingList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListBuggyFeeSetting{}
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

	buggyRequest := models.BuggyFeeSetting{}
	buggyRequest.CourseUid = form.CourseUid
	buggyRequest.PartnerUid = form.PartnerUid

	list, total, err := buggyRequest.FindList(db, page)

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
func (_ *CBuggyFeeSetting) UpdateBuggyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || Id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	body := models.BuggyFeeSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	buggyFeeSetting := models.BuggyFeeSetting{}
	buggyFeeSetting.Id = Id
	errF := buggyFeeSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	if body.Status != "" {
		buggyFeeSetting.Status = body.Status
	}

	errUpd := buggyFeeSetting.Update(db)
	if errUpd != nil {
		response_message.InternalServerError(c, errUpd.Error())
		return
	}

	okResponse(c, buggyFeeSetting)
}
func (_ *CBuggyFeeSetting) DeleteBuggyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || Id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	buggyFeeSetting := models.BuggyFeeSetting{}
	buggyFeeSetting.Id = Id
	errF := buggyFeeSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := buggyFeeSetting.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
