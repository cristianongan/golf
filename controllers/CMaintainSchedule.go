package controllers

import (
	"fmt"
	"log"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils/response_message"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nleeper/goment"
	"gorm.io/datatypes"
)

type CMaintainSchedule struct {
}

func (_ *CMaintainSchedule) CreateMaintainSchedule(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateMaintainScheduleBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateMaintainSchedule BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	hasError := false

	for _, item := range body.MaintainScheduleList {
		// validate week_id
		g, _ := goment.New(time.Now().UnixNano())
		totalWeekInYear := int64(g.WeekYear())

		weekId := strings.Split(item.WeekId, "-")
		weekIdYear, _ := strconv.ParseInt(weekId[0], 10, 64)
		weekIdWeek, _ := strconv.ParseInt(weekId[1], 10, 64)

		if int64(g.Year()) != weekIdYear {
			response_message.BadRequest(c, "week_id_year is invalid")
			hasError = true
			break
		}

		if weekIdWeek > totalWeekInYear {
			response_message.BadRequest(c, "week_id_week is invalid")
			hasError = true
			break
		}

		//TODO: validate apply_date

		hasError2 := false

		for _, applyDayOff := range item.ApplyDayOffList {
			applyDate, _ := time.Parse("2006-01-02", applyDayOff.ApplyDate)
			applyDate2 := datatypes.Date(applyDate)
			maintainSchedule := models.MaintainSchedule{
				CourseName:       item.CourseName,
				WeekId:           item.WeekId,
				ApplyDate:        &(applyDate2),
				PartnerUid:       prof.PartnerUid,
				CourseUid:        prof.CourseUid,
				MorningOff:       applyDayOff.MorningOff,
				AfternoonOff:     applyDayOff.AfternoonOff,
				MorningTimeOff:   applyDayOff.MorningTimeOff,
				AfternoonTimeOff: applyDayOff.AfternoonTimeOff,
			}

			if err := maintainSchedule.Create(db); err != nil {
				response_message.InternalServerError(c, err.Error())
				hasError2 = true
				break
			}
		}

		if hasError2 {
			log.Print("CreateMaintainSchedule.Create()")
			hasError = true
			break
		}
	}

	if hasError {
		return
	}

	okRes(c)
}

func (_ CMaintainSchedule) GetMaintainScheduleList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var query request.GetMaintainScheduleList
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

	maintainSchedule := models.MaintainSchedule{}
	maintainSchedule.PartnerUid = prof.PartnerUid
	maintainSchedule.CourseUid = prof.CourseUid
	maintainSchedule.WeekId = query.WeekId

	list, total, err := maintainSchedule.FindList(db, page)

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

func (_ *CMaintainSchedule) UpdateMaintainSchedule(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.UpdateMaintainScheduleBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateMaintainSchedule BindJSON error")
		response_message.BadRequest(c, "")
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

	hasError := false

	for _, applyDayOff := range body.ApplyDayOffList {
		applyDate, _ := time.Parse("2006-01-02", applyDayOff.ApplyDate)
		applyDate2 := datatypes.Date(applyDate)

		maintainSchedule := models.MaintainSchedule{
			CourseName: body.CourseName,
			WeekId:     body.WeekId,
			ApplyDate:  &applyDate2,
			PartnerUid: prof.PartnerUid,
		}

		if err := maintainSchedule.FindFirst(db); err != nil {

		}

		maintainSchedule.MorningOff = applyDayOff.MorningOff
		maintainSchedule.AfternoonOff = applyDayOff.AfternoonOff
		maintainSchedule.MorningTimeOff = applyDayOff.MorningTimeOff
		maintainSchedule.AfternoonTimeOff = applyDayOff.AfternoonTimeOff

		if err := maintainSchedule.Update(db); err != nil {
			response_message.InternalServerError(c, err.Error())
			hasError = true
			break
		}
	}

	if hasError {
		log.Print("CreateMaintainSchedule.Create()")
		return
	}

	okRes(c)
}
