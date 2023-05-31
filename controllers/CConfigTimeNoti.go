package controllers

import (
	"errors"
	"start/constants"
	"start/controllers/request"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CConfigTimeNoti struct{}

func (_ *CConfigTimeNoti) CreateConfig(c *gin.Context, prof models.CmsUser) {
	body := models.ConfigTimeNoti{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if message := validateRequest(body, prof); message != "" {
		badRequest(c, message)
		return
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	if body.Id != 0 {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	// create config
	model := models.ConfigTimeNoti{}

	model.TimeIntervalType = body.TimeIntervalType
	model.FirstMilestone = body.FirstMilestone
	model.SecondMilestone = body.SecondMilestone

	if model.FindFirst(db) == nil {
		response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	if body.Status == "" {
		model.Status = constants.CONFIG_TIME_NOTI_ACTIVE
	}

	model.ColorCode = body.ColorCode
	model.Description = body.Description
	model.PartnerUid = body.PartnerUid
	model.CourseUid = body.CourseUid

	errC := model.Create(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, model)
}

func (_ *CConfigTimeNoti) UpdateConfig(c *gin.Context, prof models.CmsUser) {
	body := models.ConfigTimeNoti{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	if message := validateRequest(body, prof); message != "" {
		badRequest(c, message)
		return
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	// update config
	model := models.ConfigTimeNoti{}
	model.Id = body.Id

	if model.Id == 0 || model.FindFirst(db) != nil {
		response_message.BadRequest(c, constants.DB_ERR_RECORD_NOT_FOUND)
		return
	}

	nonUid := models.ConfigTimeNoti{}
	nonUid.TimeIntervalType = body.TimeIntervalType
	nonUid.FirstMilestone = body.FirstMilestone
	nonUid.SecondMilestone = body.SecondMilestone

	if nonUid.FindFirstExclude([]int64{model.Id}, db) == nil {
		response_message.BadRequest(c, constants.API_ERR_DUPLICATED_RECORD)
		return
	}

	model.TimeIntervalType = body.TimeIntervalType
	model.FirstMilestone = body.FirstMilestone
	model.SecondMilestone = body.SecondMilestone
	model.ColorCode = body.ColorCode
	model.Description = body.Description
	model.PartnerUid = body.PartnerUid
	model.CourseUid = body.CourseUid

	errC := model.Update(db)

	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, model)
}

func validateRequest(data models.ConfigTimeNoti, prof models.CmsUser) string {
	if prof.RoleId != -1 && (prof.PartnerUid != data.PartnerUid || prof.CourseUid != data.CourseUid) {
		return constants.API_ERR_INVALID_BODY_DATA
	}

	if len(data.ColorCode) > 50 {
		return constants.API_ERR_INVALID_BODY_DATA
	}

	if len(data.Description) > 255 {
		return constants.API_ERR_INVALID_BODY_DATA
	}

	if !(data.TimeIntervalType == constants.CONFIG_TIME_NOTI_GREATER_THAN || data.TimeIntervalType == constants.CONFIG_TIME_NOTI_RANGE || data.TimeIntervalType == constants.CONFIG_TIME_NOTI_SMALLER_THAN) {
		return constants.API_ERR_INVALID_BODY_DATA
	}

	if data.FirstMilestone < 0 || data.SecondMilestone < 0 {
		return constants.CONFIG_MILE_STONE_CAN_NOT_BE_NAGATIVE
	}

	if data.TimeIntervalType == constants.CONFIG_TIME_NOTI_GREATER_THAN && data.FirstMilestone != 0 {
		return constants.CONFIG_FIRST_MILE_STONE_IS_INVALID
	} else if data.TimeIntervalType == constants.CONFIG_TIME_NOTI_SMALLER_THAN && data.SecondMilestone != 0 {
		return constants.CONFIG_SECOND_MILE_STONE_IS_INVALID
	} else if data.TimeIntervalType == constants.CONFIG_TIME_NOTI_RANGE && data.FirstMilestone >= data.SecondMilestone {
		return constants.CONFIG_FIRST_MILE_STONE_CAN_NOT_GREATER_THAN_SECOND
	}

	return ""
}

func (_ *CConfigTimeNoti) GetListConfig(c *gin.Context, prof models.CmsUser) {
	form := request.GetConfigTimeNoti{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if strings.TrimSpace(form.PartnerUid) == "" || strings.TrimSpace(form.CourseUid) == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	page := models.Page{
		Limit:   form.PageRequest.Limit,
		Page:    form.PageRequest.Page,
		SortBy:  form.PageRequest.SortBy,
		SortDir: form.PageRequest.SortDir,
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	model := models.ConfigTimeNoti{
		PartnerUid: form.PartnerUid,
	}

	if form.Status != "" {
		model.Status = form.Status
	}

	if prof.RoleId != -1 {
		// not root
		model.CourseUid = prof.CourseUid
	}

	list, total, err := model.FindList(page, db)
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

func (_ *CConfigTimeNoti) GetListConfigAvailable(c *gin.Context, prof models.CmsUser) {
	form := request.GetConfigTimeNoti{}
	if bindErr := c.ShouldBind(&form); bindErr != nil {
		response_message.BadRequest(c, bindErr.Error())
		return
	}

	if strings.TrimSpace(form.PartnerUid) == "" || strings.TrimSpace(form.CourseUid) == "" {
		response_message.BadRequest(c, constants.API_ERR_INVALID_BODY_DATA)
		return
	}

	course := models.Course{}

	course.Uid = form.CourseUid

	if cErr := course.FindFirst(); cErr != nil {
		response_message.BadRequest(c, constants.DB_ERR_RECORD_NOT_FOUND)
		return
	}

	if !course.ConfigTimeNoti {
		okResponse(c, []models.ConfigTimeNoti{})
		return
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	model := models.ConfigTimeNoti{}
	model.PartnerUid = form.PartnerUid
	model.Status = constants.CONFIG_TIME_NOTI_ACTIVE

	if prof.RoleId != -1 {
		model.CourseUid = form.CourseUid
	}

	list, total, err := model.FindAll(db)
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

func (_ *CConfigTimeNoti) DeleteConfig(c *gin.Context, prof models.CmsUser) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	course := models.ConfigTimeNoti{}
	course.Id = id
	course.CourseUid = prof.CourseUid
	course.PartnerUid = prof.PartnerUid
	errF := course.FindFirst(db)
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := course.Delete(db)
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
