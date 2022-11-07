package controllers

import (
	"errors"
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CCaddieWorkingCalendar struct{}

func (_ *CCaddieWorkingCalendar) CreateCaddieWorkingCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateCaddieWorkingCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieWorkingCalendar BindJSON error", err)
		response_message.BadRequest(c, "")
		return
	}

	now := time.Now()
	listCreate := []models.CaddieWorkingCalendar{}

	for _, v := range body.CaddieList {
		caddieWC := models.CaddieWorkingCalendar{}
		caddieWC.CreatedAt = now.Unix()
		caddieWC.UpdatedAt = now.Unix()
		caddieWC.Status = constants.STATUS_ENABLE
		caddieWC.PartnerUid = v.PartnerUid
		caddieWC.CourseUid = v.CourseUid
		caddieWC.CaddieCode = v.CaddieCode
		caddieWC.ApplyDate = v.ApplyDate
		caddieWC.NumberOrder = v.NumberOrder
		listCreate = append(listCreate, caddieWC)
	}

	// validate caddie_uid
	caddieWC := models.CaddieWorkingCalendar{}

	if err := caddieWC.BatchInsert(db, listCreate); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieWorkingCalendar) GetCaddieWorkingCalendarList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	// TODO: filter by from and to

	body := request.GetCaddieWorkingCalendarList{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieWorkingCalendar := models.CaddieWorkingCalendar{}
	caddieWorkingCalendar.CourseUid = body.CourseUid
	caddieWorkingCalendar.PartnerUid = body.PartnerUid
	caddieWorkingCalendar.ApplyDate = body.ApplyDate

	list, total, err := caddieWorkingCalendar.FindAllByDate(db)

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

func (_ *CCaddieWorkingCalendar) UpdateCaddieWorkingCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	caddeWCIdStr := c.Param("id")
	caddeWCId, err := strconv.ParseInt(caddeWCIdStr, 10, 64)
	if err != nil || caddeWCId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	var body request.UpdateCaddieWorkingCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieWorkingCalendar BindJSON error")
		response_message.BadRequest(c, "")
	}

	caddiWC := models.CaddieWorkingCalendar{}
	caddiWC.Id = caddeWCId

	if err := caddiWC.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddiWC.CaddieCode = body.CaddieCode

	if err := caddiWC.Update(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieWorkingCalendar) DeleteCaddieWorkingCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	caddeWCIdStr := c.Param("id")
	caddeWCId, err := strconv.ParseInt(caddeWCIdStr, 10, 64)
	if err != nil || caddeWCId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	caddiWC := models.CaddieWorkingCalendar{}
	caddiWC.Id = caddeWCId

	if err := caddiWC.Delete(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	okRes(c)
}
