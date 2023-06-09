package controllers

import (
	"log"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/datatypes"
)

type CCaddieCalendar struct{}

func (_ *CCaddieCalendar) CreateCaddieCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateCaddieCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	fromDate, err := time.Parse("2006-01-02", body.FromDate)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	toDate, err := time.Parse("2006-01-02", body.ToDate)

	if err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate fromDate and toDate
	if fromDate.Format("2006-01") != toDate.Format("2006-01") {
		response_message.BadRequest(c, "From date and to date do not have same month")
		return
	}

	var caddieCalendars []models.CaddieCalendar

	for _, caddieUid := range body.CaddieUidList {
		// validate caddie_uid
		caddie := models.Caddie{}
		caddie.Id = caddieUid

		if err := caddie.FindFirst(db); err != nil {
			response_message.BadRequest(c, err.Error())
			return
		}

		dateRange := utils.DateRangeNew(fromDate, toDate, utils.DAYS)

		for dateRange.Next() {
			applyDate := dateRange.Current()

			// validate apply_date duplicate
			caddieCalendarList := models.CaddieCalendarList{}
			caddieCalendarList.ApplyDate = applyDate.Format("2006-01-02")
			caddieCalendarList.CaddieCode = caddie.Code
			if body.DayOffType != constants.DAY_OFF_TYPE_SICK {
				caddieCalendarList.DayOffType = body.DayOffType
			}
			if _, err := caddieCalendarList.FindFirst(); err == nil {
				response_message.BadRequest(c, "record duplicate"+" ["+applyDate.Format("2006-01-02")+"]")
				return
			}

			caddieCalendar := models.CaddieCalendar{
				CaddieUid:  strconv.FormatInt(caddie.Id, 10),
				CaddieCode: caddie.Code,
				CaddieName: caddie.Name,
				PartnerUid: caddie.PartnerUid,
				CourseUid:  caddie.CourseUid,
				Title:      body.Title,
				DayOffType: body.DayOffType,
				ApplyDate:  datatypes.Date(applyDate),
				Note:       body.Note,
			}

			if err := caddieCalendar.Create(db); err != nil {
				log.Print("CCaddieCalendar.Create()")
				response_message.InternalServerError(c, err.Error())
				return
			}

			caddieCalendars = append(caddieCalendars, caddieCalendar)
		}
	}

	c.JSON(200, caddieCalendars)
}

func (_ *CCaddieCalendar) GetCaddieCalendarList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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

	caddie := models.CaddieList{}
	caddie.CourseUid = prof.CourseUid
	caddie.CaddieName = query.CaddieName
	caddie.CaddieCode = query.CaddieCode
	caddie.Month = query.Month

	list, total, err := caddie.FindList(db, page)

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
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	caddieCalendar.PartnerUid = prof.PartnerUid
	caddieCalendar.CourseUid = prof.CourseUid

	if err := caddieCalendar.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieCalendar.Title = body.Title
	caddieCalendar.DayOffType = body.DayOffType
	caddieCalendar.Note = body.Note

	if err := caddieCalendar.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieCalendar) DeleteMonthCaddieCalendar(c *gin.Context, prof models.CmsUser) {
	var body request.DeleteMonthCaddieCalendarBody

	if err := c.BindJSON(&body); err != nil {
		log.Print("DeleteMonthCaddieCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	for _, caddieUid := range body.CaddieUidList {
		caddieCalendarList := models.CaddieCalendarList{}
		caddieCalendarList.CourseUid = prof.CourseUid
		caddieCalendarList.CaddieUid = strconv.FormatInt(caddieUid, 10)
		caddieCalendarList.Month = body.Month

		if err := caddieCalendarList.Delete(); err != nil {
			response_message.InternalServerError(c, err.Error())
			return
		}
	}

	okRes(c)
}

func (_ *CCaddieCalendar) DeleteDateCaddieCalendar(c *gin.Context, prof models.CmsUser) {
	var body request.DeleteDateCaddieCalendarBody
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	if err := c.BindJSON(&body); err != nil {
		log.Print("DeleteDateCaddieCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	applyDate, _ := time.Parse("2006-01-02", body.Date)

	caddieCalendar := models.CaddieCalendar{}
	caddieCalendar.CourseUid = prof.CourseUid
	caddieCalendar.CaddieUid = strconv.FormatInt(body.CaddieUid, 10)
	caddieCalendar.ApplyDate = datatypes.Date(applyDate)
	if body.DayOffType != "" && body.DayOffType != constants.DAY_OFF_TYPE_SICK {
		caddieCalendar.DayOffType = body.DayOffType
	}
	if err := caddieCalendar.FindFirst(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	if err := caddieCalendar.Delete(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}
