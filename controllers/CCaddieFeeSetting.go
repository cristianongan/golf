package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CCaddieFeeSetting struct{}

// / --------- CaddieFee Setting Group ----------
func (_ *CCaddieFeeSetting) CreateCaddieFeeSettingGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.CaddieFeeSettingGroup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated(db) {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	CaddieFeeSettingGroup := models.CaddieFeeSettingGroup{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		Name:       body.Name,
		FromDate:   body.FromDate,
		ToDate:     body.ToDate,
	}

	errC := CaddieFeeSettingGroup.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, CaddieFeeSettingGroup)
}

func (_ *CCaddieFeeSetting) GetListCaddieFeeSettingGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListCaddieFeeSettingGroupForm{}
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

	CaddieFeeSettingGroupR := models.CaddieFeeSettingGroup{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}
	list, total, err := CaddieFeeSettingGroupR.FindList(db, page, 0, 0)
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

func (_ *CCaddieFeeSetting) UpdateCaddieFeeSettingGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	CaddieFeeSettingGroupIdStr := c.Param("id")
	CaddieFeeSettingGroupId, err := strconv.ParseInt(CaddieFeeSettingGroupIdStr, 10, 64)
	if err != nil || CaddieFeeSettingGroupId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	CaddieFeeSettingGroup := models.CaddieFeeSettingGroup{}
	CaddieFeeSettingGroup.Id = CaddieFeeSettingGroupId
	CaddieFeeSettingGroup.PartnerUid = prof.PartnerUid
	CaddieFeeSettingGroup.CourseUid = prof.CourseUid
	errF := CaddieFeeSettingGroup.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.CaddieFeeSettingGroup{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if CaddieFeeSettingGroup.Name != body.Name || CaddieFeeSettingGroup.FromDate != body.FromDate || CaddieFeeSettingGroup.ToDate != body.ToDate {
		if body.IsDuplicated(db) {
			response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
			return
		}
	}

	if body.Name != "" {
		CaddieFeeSettingGroup.Name = body.Name
	}
	if body.Status != "" {
		CaddieFeeSettingGroup.Status = body.Status
	}
	if body.FromDate != 0 {
		CaddieFeeSettingGroup.FromDate = body.FromDate
	}
	if body.ToDate != 0 {
		CaddieFeeSettingGroup.ToDate = body.ToDate
	}

	errUdp := CaddieFeeSettingGroup.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, CaddieFeeSettingGroup)
}

func (_ *CCaddieFeeSetting) DeleteCaddieFeeSettingGroup(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	CaddieFeeSettingGroupIdStr := c.Param("id")
	CaddieFeeSettingGroupId, err := strconv.ParseInt(CaddieFeeSettingGroupIdStr, 10, 64)
	if err != nil || CaddieFeeSettingGroupId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	CaddieFeeSettingGroup := models.CaddieFeeSettingGroup{}
	CaddieFeeSettingGroup.Id = CaddieFeeSettingGroupId
	CaddieFeeSettingGroup.PartnerUid = prof.PartnerUid
	CaddieFeeSettingGroup.CourseUid = prof.CourseUid
	errF := CaddieFeeSettingGroup.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := CaddieFeeSettingGroup.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}

/// --------- CaddieFee Setting ----------

func (_ *CCaddieFeeSetting) CreateCaddieFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.CaddieFeeSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if !body.IsValidated() {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	if body.IsDuplicated(db) {
		response_message.DuplicateRecord(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	//Check Group Id avaible
	caddieFeeSettingGroup := models.CaddieFeeSettingGroup{}
	caddieFeeSettingGroup.Id = body.GroupId
	errFind := caddieFeeSettingGroup.FindFirst(db)
	if errFind != nil {
		response_message.BadRequest(c, errFind.Error())
		return
	}

	CaddieFeeSetting := models.CaddieFeeSetting{
		PartnerUid: body.PartnerUid,
		CourseUid:  body.CourseUid,
		GroupId:    body.GroupId,
		Type:       body.Type,
		Fee:        body.Fee,
		Hole:       body.Hole,
	}

	CaddieFeeSetting.Status = body.Status

	errC := CaddieFeeSetting.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, CaddieFeeSetting)
}

func (_ *CCaddieFeeSetting) GetListCaddieFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListCaddieFeeSettingForm{}
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

	CaddieFeeSettingR := models.CaddieFeeSetting{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
		GroupId:    form.GroupId,
	}
	list, total, err := CaddieFeeSettingR.FindList(db, page)
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

func (_ *CCaddieFeeSetting) UpdateCaddieFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	CaddieFeeSettingIdStr := c.Param("id")
	CaddieFeeSettingId, err := strconv.ParseInt(CaddieFeeSettingIdStr, 10, 64)
	if err != nil || CaddieFeeSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	CaddieFeeSetting := models.CaddieFeeSetting{}
	CaddieFeeSetting.Id = CaddieFeeSettingId
	CaddieFeeSetting.PartnerUid = prof.PartnerUid
	CaddieFeeSetting.CourseUid = prof.CourseUid
	errF := CaddieFeeSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.CaddieFeeSetting{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Status != "" {
		CaddieFeeSetting.Status = body.Status
	}

	CaddieFeeSetting.Type = body.Type
	CaddieFeeSetting.Hole = body.Hole
	CaddieFeeSetting.Fee = body.Fee

	errUdp := CaddieFeeSetting.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, CaddieFeeSetting)
}

func (_ *CCaddieFeeSetting) DeleteCaddieFeeSetting(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	CaddieFeeSettingIdStr := c.Param("id")
	CaddieFeeSettingId, err := strconv.ParseInt(CaddieFeeSettingIdStr, 10, 64)
	if err != nil || CaddieFeeSettingId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	CaddieFeeSetting := models.CaddieFeeSetting{}
	CaddieFeeSetting.Id = CaddieFeeSettingId
	CaddieFeeSetting.PartnerUid = prof.PartnerUid
	CaddieFeeSetting.CourseUid = prof.CourseUid
	errF := CaddieFeeSetting.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := CaddieFeeSetting.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
