package controllers

import (
	"errors"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CGroupFee struct{}

func (_ *CGroupFee) GetListGroupFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	form := request.GetListGroupFeeForm{}
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

	groupFeeR := models.GroupFee{
		PartnerUid: form.PartnerUid,
		CourseUid:  form.CourseUid,
	}

	groupFeeR.Status = form.Status
	list, total, err := groupFeeR.FindList(db, page)
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

func (_ *CGroupFee) CreateGroupFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	body := models.GroupFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if body.PartnerUid == "" || body.CourseUid == "" {
		response_message.BadRequest(c, "data not valid")
		return
	}

	groupFee := models.GroupFee{}
	groupFee.PartnerUid = body.PartnerUid
	groupFee.CourseUid = body.CourseUid
	groupFee.Name = body.Name
	groupFee.CategoryType = body.CategoryType

	// Check duplicated
	errF := groupFee.FindFirst(db)
	if errF == nil || groupFee.Id > 0 {
		response_message.BadRequest(c, errF.Error())
		return
	}

	groupFee.Status = body.Status

	errC := groupFee.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, groupFee)
}

func (_ *CGroupFee) UpdateGroupFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	groupFeeIdStr := c.Param("id")
	groupFeeId, err := strconv.ParseInt(groupFeeIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && groupFeeId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	groupFee := models.GroupFee{}
	groupFee.Id = groupFeeId
	errF := groupFee.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	body := models.GroupFee{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if body.Name != "" {
		groupFee.Name = body.Name
	}
	if body.Status != "" {
		groupFee.Status = body.Status
	}
	if body.CategoryType != "" {
		groupFee.CategoryType = body.CategoryType
	}

	errUdp := groupFee.Update(db)
	if errUdp != nil {
		response_message.InternalServerError(c, errUdp.Error())
		return
	}

	okResponse(c, groupFee)
}

func (_ *CGroupFee) DeleteGroupFee(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	groupFeeIdStr := c.Param("id")
	groupFeeId, err := strconv.ParseInt(groupFeeIdStr, 10, 64) // Nếu uid là int64 mới cần convert
	if err != nil && groupFeeId == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	groupFee := models.GroupFee{}
	groupFee.Id = groupFeeId
	errF := groupFee.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := groupFee.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
