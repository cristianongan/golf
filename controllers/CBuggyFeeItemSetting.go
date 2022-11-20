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

type CBuggyFeeItemSetting struct{}

func (_ *CBuggyFeeItemSetting) CreateBuggyFeeItemSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.BuggyFeeItemSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	buggyFeeSetting := models.BuggyFeeSetting{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		ModelId:    models.ModelId{Id: body.SettingId},
	}

	if errFind := buggyFeeSetting.FindFirst(db); errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	buggyFeeItemSetting := models.BuggyFeeItemSetting{
		PartnerUid:    body.PartnerUid,
		CourseUid:     body.CourseUid,
		GuestStyle:    body.GuestStyle,
		Dow:           body.Dow,
		RentalFee:     body.RentalFee,
		PrivateCarFee: body.PrivateCarFee,
		OddCarFee:     body.OddCarFee,
		SettingId:     body.SettingId,
		RateGolfFee:   body.RateGolfFee,
	}
	buggyFeeItemSetting.Status = body.Status
	errC := buggyFeeItemSetting.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, buggyFeeItemSetting)
}

func (_ *CBuggyFeeItemSetting) GetBuggyFeeItemSettingList(c *gin.Context, prof models.CmsUser) {
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

	buggyRequest := models.BuggyFeeItemSetting{}
	buggyRequest.CourseUid = form.CourseUid
	buggyRequest.PartnerUid = form.PartnerUid
	buggyRequest.SettingId = form.SettingId

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

func (_ *CBuggyFeeItemSetting) UpdateBuggyFeeItemSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || Id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	body := models.BuggyFeeItemSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	buggyFeeItemSetting := models.BuggyFeeItemSetting{}
	buggyFeeItemSetting.Id = Id
	errF := buggyFeeItemSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}
	if body.GuestStyle != "" {
		buggyFeeItemSetting.GuestStyle = body.GuestStyle
	}
	if body.Dow != "" {
		buggyFeeItemSetting.Dow = body.Dow
	}
	if body.RateGolfFee != "" {
		buggyFeeItemSetting.RateGolfFee = body.RateGolfFee
	}
	if body.OddCarFee != nil {
		buggyFeeItemSetting.OddCarFee = body.OddCarFee
	}
	if body.PrivateCarFee != nil {
		buggyFeeItemSetting.PrivateCarFee = body.PrivateCarFee
	}
	if body.RentalFee != nil {
		buggyFeeItemSetting.RentalFee = body.RentalFee
	}
	if body.Status != "" {
		buggyFeeItemSetting.Status = body.Status
	}

	errUpd := buggyFeeItemSetting.Update(db)
	if errUpd != nil {
		response_message.InternalServerError(c, errUpd.Error())
		return
	}

	okResponse(c, buggyFeeItemSetting)
}

func (_ *CBuggyFeeItemSetting) DeleteBuggyFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	IdStr := c.Param("id")
	Id, err := strconv.ParseInt(IdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil || Id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	buggyFeeSetting := models.BuggyFeeItemSetting{}
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
