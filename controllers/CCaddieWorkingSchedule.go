package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nleeper/goment"
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

type CCaddieWorkingSchedule struct {
}

func (_ *CCaddieWorkingSchedule) CreateCaddieWorkingSchedule(c *gin.Context, prof models.CmsUser) {
	var body request.CreateWorkingScheduleBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieWorkingSchedule BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	// validate caddie_group
	caddieGroup := models.CaddieGroup{}
	caddieGroup.Code = body.CaddieGroupCode
	if err := caddieGroup.FindFirst(); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate week_id
	g, _ := goment.New(time.Now().UnixNano())
	totalWeekInYear := int64(g.WeekYear())

	weekId := strings.Split(body.WeekId, "-")
	weekIdYear, _ := strconv.ParseInt(weekId[0], 10, 64)
	weekIdWeek, _ := strconv.ParseInt(weekId[1], 10, 64)

	if int64(g.Year()) != weekIdYear {
		response_message.BadRequest(c, "week_id_year is invalid")
		return
	}

	if weekIdWeek > totalWeekInYear {
		response_message.BadRequest(c, "week_id_week is invalid")
		return
	}

	//TODO: validate apply_date

	hasError := false

	for _, applyDayOff := range body.ApplyDayOffList {
		applyDate, _ := time.Parse("2006-01-02", applyDayOff.ApplyDate)
		caddieWorkingSchedule := models.CaddieWorkingSchedule{
			CaddieGroupName: caddieGroup.Name,
			CaddieGroupCode: caddieGroup.Code,
			WeekId:          body.WeekId,
			ApplyDate:       datatypes.Date(applyDate),
			IsDayOff:        applyDayOff.IsDayOff,
			PartnerUid:      prof.PartnerUid,
			CourseUid:       prof.CourseUid,
		}

		if err := caddieWorkingSchedule.Create(); err != nil {
			response_message.InternalServerError(c, err.Error())
			hasError = true
			break
		}
	}

	if hasError {
		log.Print("CreateCaddieWorkingSchedule.Create()")
		return
	}

	okRes(c)
}

func (_ CCaddieWorkingSchedule) GetCaddieWorkingScheduleList(c *gin.Context, prof models.CmsUser) {
	var query request.GetCaddieWorkingScheduleList
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

	fmt.Println("[DEBUG]", query.WeekId)

	caddieWorkingSchedule := models.CaddieWorkingSchedule{}
	caddieWorkingSchedule.PartnerUid = prof.PartnerUid
	caddieWorkingSchedule.CourseUid = prof.CourseUid
	caddieWorkingSchedule.WeekId = query.WeekId

	list, total, err := caddieWorkingSchedule.FindList(page)

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
