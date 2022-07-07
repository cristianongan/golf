package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"log"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"time"
)

type CCaddieCalendar struct{}

func (_ *CCaddieCalendar) CreateCaddieCalendar(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	applyDate, _ := time.Parse("2006-01-02", body.ApplyDate)

	// validate caddie_uid
	caddie := models.Caddie{}
	caddie.Id, _ = strconv.ParseInt(body.CaddieUid, 10, 64)
	if err := caddie.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate apply_date duplicate
	caddieCalendarList := models.CaddieCalendarList{}
	caddieCalendarList.ApplyDate = applyDate.Format("2006-01-02")
	if _, err := caddieCalendarList.FindFirst(); err == nil {
		response_message.BadRequest(c, "record duplicate")
		return
	}

	caddieCalendar := models.CaddieCalendar{
		CaddieUid:  body.CaddieUid,
		CaddieCode: caddie.Code,
		CaddieName: caddie.Name,
		PartnerUid: caddie.PartnerUid,
		CourseUid:  caddie.CourseUid,
		Title:      body.Title,
		DayOffType: body.DayOffType,
		ApplyDate:  datatypes.Date(applyDate),
		Note:       body.Note,
	}

	if err := caddieCalendar.Create(); err != nil {
		log.Print("CCaddieCalendar.Create()")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieCalendar)
}

func (_ *CCaddieCalendar) GetCaddieCalendarList(c *gin.Context, prof models.CmsUser) {
	// TODO: filter by month

	query := request.GetCaddieCalendarList{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   query.PageRequest.Limit,
		Page:    query.PageRequest.Page,
		SortBy:  query.PageRequest.SortBy,
		SortDir: query.PageRequest.SortDir,
	}

	//caddieCalendar := models.CaddieCalendarList{}
	//
	//caddieCalendar.CourseUid = prof.CourseUid
	//caddieCalendar.CaddieName = query.CaddieName
	//caddieCalendar.CaddieCode = query.CaddieCode
	//caddieCalendar.Month = query.Month
	//
	//list, total, err := caddieCalendar.FindList(page)

	caddie := models.CaddieList{}
	caddie.CourseUid = prof.CourseUid
	caddie.CaddieName = query.CaddieName
	caddie.CaddieCode = query.CaddieCode
	caddie.Month = query.Month

	list, total, err := caddie.FindList(page)

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

func (_ *CCaddieCalendar) UpdateCaddieCalendar(c *gin.Context, prof models.CmsUser) {
	var body request.UpdateCaddieCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	applyDate, _ := time.Parse("2006-01-02", body.ApplyDate)

	caddieCalendar := models.CaddieCalendar{}
	caddieCalendar.CaddieUid = body.CaddieUid
	caddieCalendar.ApplyDate = datatypes.Date(applyDate)
	caddieCalendar.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)

	if err := caddieCalendar.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieCalendar.Title = body.Title
	caddieCalendar.DayOffType = body.DayOffType
	caddieCalendar.Note = body.Note

	if err := caddieCalendar.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
