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
	"strings"
	"time"
)

type CCaddieWorkingCalendar struct{}

func (_ *CCaddieWorkingCalendar) CreateCaddieWorkingCalendar(c *gin.Context, prof models.CmsUser) {
	var body request.CreateCaddieWorkingCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieWorkingCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	// validate caddie_uid
	caddie := models.Caddie{}
	caddie.Id, _ = strconv.ParseInt(body.CaddieUid, 10, 64)
	if err := caddie.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// TODO: validate row_time + apply_date + caddie_column + caddie_row

	applyDate, _ := time.Parse("2006-01-02", body.ApplyDate)

	rowTime := strings.Split(body.RowTime, ":")

	rowTimeHour, _ := strconv.ParseInt(rowTime[0], 10, 32)

	rowTimeMinute, _ := strconv.ParseInt(rowTime[1], 10, 32)

	caddieColumn, _ := strconv.ParseInt(body.CaddieColumn, 10, 32)

	caddieWorkingCalendar := models.CaddieWorkingCalendar{
		CaddieUid:    body.CaddieUid,
		CaddieCode:   caddie.Code,
		PartnerUid:   prof.PartnerUid,
		CourseUid:    prof.CourseUid,
		CaddieLabel:  body.CaddieLabel,
		CaddieColumn: int(caddieColumn),
		CaddieRow:    body.CaddieRow,
		RowTime:      datatypes.NewTime(int(rowTimeHour), int(rowTimeMinute), 0, 0),
		ApplyDate:    datatypes.Date(applyDate),
	}

	if err := caddieWorkingCalendar.Create(); err != nil {
		log.Print("CreateCaddieWorkingCalendar.Create()")
		response_message.InternalServerError(c, err.Error())
		return
	}
	c.JSON(200, caddieWorkingCalendar)
}

func (_ *CCaddieWorkingCalendar) GetCaddieWorkingCalendarList(c *gin.Context, prof models.CmsUser) {
	// TODO: filter by from and to

	query := request.GetCaddieWorkingCalendarList{}
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

	caddieWorkingCalendar := models.CaddieWorkingCalendarList{}
	caddieWorkingCalendar.CourseUid = prof.CourseUid
	caddieWorkingCalendar.ApplyDate = query.ApplyDate

	list, total, err := caddieWorkingCalendar.FindList(page)

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
	var body request.UpdateCaddieWorkingCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieWorkingCalendar BindJSON error")
		response_message.BadRequest(c, "")
	}

	// validate caddie_uid
	caddie := models.Caddie{}
	caddie.Id, _ = strconv.ParseInt(body.CaddieUid, 10, 64)
	caddie.Code = body.CaddieCode
	if err := caddie.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	applyDate, _ := time.Parse("2006-01-02", body.ApplyDate)

	rowTime := strings.Split(body.RowTime, ":")

	rowTimeHour, _ := strconv.ParseInt(rowTime[0], 10, 32)

	rowTimeMinute, _ := strconv.ParseInt(rowTime[1], 10, 32)

	caddieColumn, _ := strconv.ParseInt(body.CaddieColumn, 10, 32)

	caddieWorkingCalendar := models.CaddieWorkingCalendar{}
	caddieWorkingCalendar.Id, _ = strconv.ParseInt(c.Param("id"), 10, 64)
	caddieWorkingCalendar.ApplyDate = datatypes.Date(applyDate)
	caddieWorkingCalendar.CaddieColumn = int(caddieColumn)
	caddieWorkingCalendar.CaddieRow = body.CaddieRow
	caddieWorkingCalendar.RowTime = datatypes.NewTime(int(rowTimeHour), int(rowTimeMinute), 0, 0)

	if err := caddieWorkingCalendar.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieWorkingCalendar.CaddieUid = body.CaddieUid
	caddieWorkingCalendar.CaddieCode = body.CaddieCode
	caddieWorkingCalendar.CaddieLabel = body.CaddieLabel

	if err := caddieWorkingCalendar.Update(); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
