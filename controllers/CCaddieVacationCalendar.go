package controllers

import (
	"encoding/json"
	"errors"
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
)

type CCaddieVacationCalendar struct{}

func (_ *CCaddieVacationCalendar) CreateCaddieVacationCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	var body request.CreateCaddieVacationCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("CreateCaddieVacationCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	validate := validator.New()

	if err := validate.Struct(body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// validate fromDate and toDate
	if body.DateFrom > body.DateTo {
		response_message.BadRequest(c, "To date must be greater than from date")
		return
	}

	// validate caddie
	caddie := models.Caddie{}

	caddie.Id = body.CaddieId
	if err := caddie.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieVC := models.CaddieVacationCalendar{
		PartnerUid:    body.PartnerUid,
		CourseUid:     body.CourseUid,
		CaddieId:      body.CaddieId,
		CaddieCode:    caddie.Code,
		CaddieName:    caddie.Name,
		Title:         body.Title,
		Color:         body.Color,
		DateFrom:      body.DateFrom,
		DateTo:        body.DateTo,
		MonthFrom:     int(time.Unix(body.DateFrom, 0).Local().Month()),
		MonthTo:       int(time.Unix(body.DateTo, 0).Local().Month()),
		NumberDayOff:  body.NumberDayOff,
		Note:          body.Note,
		ApproveStatus: constants.CADDIE_VACATION_PENDING,
	}

	if err := caddieVC.Create(db); err != nil {
		log.Print("Craete caddie vacation calendar ")
		response_message.InternalServerError(c, err.Error())
		return
	}

	go func() {
		cNotification := CNotification{}
		cNotification.CreateCaddieVacationNotification(db, request.GetCaddieVacationNotification{
			Caddie:       caddie,
			DateFrom:     body.DateFrom,
			DateTo:       body.DateTo,
			NumberDayOff: body.NumberDayOff,
			Title:        body.Title,
			CreateAt:     caddieVC.CreatedAt,
			UserName:     prof.UserName,
			Id:           caddieVC.Id,
		})
	}()

	c.JSON(200, caddieVC)
}

func (_ *CCaddieVacationCalendar) GetCaddieVacationCalendarList(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)

	body := request.GetCaddieVacationCalendarList{}
	if err := c.Bind(&body); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	page := models.Page{
		Limit:   body.PageRequest.Limit,
		Page:    body.PageRequest.Page,
		SortBy:  body.PageRequest.SortBy,
		SortDir: body.PageRequest.SortDir,
	}

	caddie := models.Caddie{}
	caddie.PartnerUid = body.PartnerUid
	caddie.CourseUid = body.CourseUid
	caddie.Name = body.CaddieName
	caddie.Code = body.CaddieCode

	list, total, err := caddie.FindList(db, page)

	if err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	listData := make([]map[string]interface{}, len(list))

	for i, data := range list {
		//find all
		caddieVC := models.CaddieVacationCalendar{}
		caddieVC.PartnerUid = body.PartnerUid
		caddieVC.CourseUid = body.CourseUid
		caddieVC.CaddieId = data.Id
		caddieVC.MonthFrom = body.Month

		listItem, errCVC := caddieVC.FindAll(db)

		if errCVC != nil {
			response_message.BadRequest(c, errCVC.Error())
			return
		}

		// Add infor to response
		listData[i] = map[string]interface{}{
			"caddie_infor": data,
			"calendar":     listItem,
		}
	}

	res := response.PageResponse{
		Total: total,
		Data:  listData,
	}

	c.JSON(200, res)
}

func (_ *CCaddieVacationCalendar) UpdateCaddieVacationCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddieVCStr := c.Param("id")
	caddieVCId, err := strconv.ParseInt(caddieVCStr, 10, 64)
	if err != nil || caddieVCId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	var body request.UpdateCaddieVacationCalendarBody
	if err := c.BindJSON(&body); err != nil {
		log.Print("UpdateCaddieVacationCalendar BindJSON error")
		response_message.BadRequest(c, "")
		return
	}

	caddieVC := models.CaddieVacationCalendar{}
	caddieVC.Id = caddieVCId
	caddieVC.PartnerUid = prof.PartnerUid
	caddieVC.CourseUid = prof.CourseUid

	if err := caddieVC.FindFirst(db); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	caddieVC.Title = body.Title
	caddieVC.Color = body.Color
	caddieVC.DateFrom = body.DateFrom
	caddieVC.DateTo = body.DateTo
	caddieVC.MonthFrom = int(time.Unix(body.DateFrom, 0).Local().Month())
	caddieVC.MonthTo = int(time.Unix(body.DateTo, 0).Local().Month())
	caddieVC.NumberDayOff = body.NumberDayOff
	caddieVC.Note = body.Note

	if err := caddieVC.Update(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}

	okRes(c)
}

func (_ *CCaddieVacationCalendar) DeleteCaddieVacationCalendar(c *gin.Context, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
	caddieVCStr := c.Param("id")
	caddieVCId, err := strconv.ParseInt(caddieVCStr, 10, 64)
	if err != nil || caddieVCId <= 0 {
		response_message.BadRequest(c, errors.New("Id not valid").Error())
		return
	}

	caddieVC := models.CaddieVacationCalendar{}
	caddieVC.Id = caddieVCId
	caddieVC.PartnerUid = prof.PartnerUid
	caddieVC.CourseUid = prof.CourseUid

	if err := caddieVC.Delete(db); err != nil {
		response_message.InternalServerError(c, err.Error())
		return
	}
	okRes(c)
}

func (_ *CCaddieVacationCalendar) UpdateCaddieVacationStatus(content []byte, isApprove bool, partnerUid string, prof models.CmsUser) {
	db := datasources.GetDatabaseWithPartner(partnerUid)

	caddieEx := models.CaddieContentNoti{}
	if err := json.Unmarshal(content, &caddieEx); err != nil {
		return
	}

	RCaddieVacation := models.CaddieVacationCalendar{
		ModelId: models.ModelId{
			Id: caddieEx.Id,
		},
	}

	if err := RCaddieVacation.FindFirst(db); err == nil {
		if isApprove {
			day := utils.GetLocalUnixTime().Unix()

			if day >= RCaddieVacation.DateFrom && day <= RCaddieVacation.DateTo {
				date, _ := utils.GetBookingDateFromTimestamp(time.Now().Unix())
				removeCaddieOutSlotOnDate(RCaddieVacation.PartnerUid, RCaddieVacation.CourseUid, date, RCaddieVacation.CaddieCode)
			}
			RCaddieVacation.ApproveStatus = constants.CADDIE_VACATION_APPROVED
		} else {
			RCaddieVacation.ApproveStatus = constants.CADDIE_VACATION_REJECTED
		}
		RCaddieVacation.ApproveTime = utils.GetTimeNow().Unix()
		RCaddieVacation.UserApprove = prof.UserName
		RCaddieVacation.Update(db)
	}
}
