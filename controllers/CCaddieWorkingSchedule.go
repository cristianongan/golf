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

	hasError := false

	for _, item := range body.CaddieGroupList {
		// validate caddie_group
		caddieGroup := models.CaddieGroup{}
		caddieGroup.Code = item.CaddieGroupCode
		if err := caddieGroup.FindFirst(); err != nil {
			response_message.BadRequest(c, err.Error())
			hasError = true
			break
		}

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
			caddieWorkingSchedule := models.CaddieWorkingSchedule{
				CaddieGroupName: caddieGroup.Name,
				CaddieGroupCode: caddieGroup.Code,
				WeekId:          item.WeekId,
				ApplyDate:       datatypes.Date(applyDate),
				IsDayOff:        applyDayOff.IsDayOff,
				PartnerUid:      prof.PartnerUid,
				CourseUid:       prof.CourseUid,
			}

			if err := caddieWorkingSchedule.Create(); err != nil {
				response_message.InternalServerError(c, err.Error())
				hasError2 = true
				break
			}
		}

		if hasError2 {
			log.Print("CreateCaddieWorkingSchedule.Create()")
			hasError = true
			break
		}
	}

	if hasError {
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

func (_ *CCaddieWorkingSchedule) UpdateCaddieWorkingSchedule(c *gin.Context, prof models.CmsUser) {
	var body request.UpdateWorkingScheduleBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieWorkingSchedule BindJSON error")
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

	hasError := false

	for _, applyDayOff := range body.ApplyDayOffList {
		applyDate, _ := time.Parse("2006-01-02", applyDayOff.ApplyDate)
		//
		//caddieWorkingSchedule := models.CaddieWorkingSchedule{
		//	CaddieGroupName: caddieGroup.Name,
		//	CaddieGroupCode: caddieGroup.Code,
		//	WeekId:          body.WeekId,
		//	ApplyDate:       datatypes.Date(applyDate),
		//	IsDayOff:        applyDayOff.IsDayOff,
		//	PartnerUid:      prof.PartnerUid,
		//	CourseUid:       prof.CourseUid,
		//}

		caddieWorkingSchedule := models.CaddieWorkingSchedule{
			CaddieGroupCode: caddieGroup.Code,
			WeekId:          body.WeekId,
			ApplyDate:       datatypes.Date(applyDate),
		}

		if err := caddieWorkingSchedule.FindFirst(); err != nil {

		}

		caddieWorkingSchedule.IsDayOff = applyDayOff.IsDayOff

		if err := caddieWorkingSchedule.Update(); err != nil {
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
