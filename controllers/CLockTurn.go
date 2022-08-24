package controllers

import (
	"errors"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"

	// "start/utils"
	"start/utils/response_message"
	"strconv"

	// "strings"
	// "time"

	"github.com/gin-gonic/gin"
)

type CLockTurn struct{}

// func (_ *CLockTurn) CreateLockTurn(c *gin.Context, prof models.CmsUser) {
// 	body := request.CreateLockTurn{}
// 	if bindErr := c.ShouldBind(&body); bindErr != nil {
// 		badRequest(c, bindErr.Error())
// 		return
// 	}

// 	course := models.Course{}
// 	course.Uid = body.CourseUid
// 	errCourse := course.FindFirst()
// 	if errCourse != nil {
// 		response_message.InternalServerError(c, errCourse.Error())
// 		return
// 	}

// 	form := request.GetListBookingSettingForm{
// 		CourseUid:  body.CourseUid,
// 		PartnerUid: body.PartnerUid,
// 	}

// 	cBookingSetting := CBookingSetting{}
// 	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(form)
// 	weekday := strconv.Itoa(int(time.Now().Weekday()))
// 	turnTimeH := 2
// 	turnLength := 0

// 	for _, data := range listSettingDetail {
// 		if strings.ContainsAny(data.Dow, weekday) {
// 			turnLength = data.TurnLength
// 			break
// 		}
// 	}

// 	listTeeTimeLock := models.ListTee{}
// 	teeTimeDate, _ := utils.ConvertHourToTime(body.TeeTime)
// 	teeList := []string{}

// 	if course.Hole == 18 {

// 		if body.Tee == "1" {
// 			teeList = []string{"10"}
// 		} else {
// 			teeList = []string{"1"}
// 		}
// 	} else if course.Hole == 27 {

// 		if body.Tee == "1A" {
// 			teeList = []string{"1B", "1C"}
// 		} else if body.Tee == "1B" {
// 			teeList = []string{"1C", "1A"}
// 		} else if body.Tee == "1C" {
// 			teeList = []string{"1A", "1B"}
// 		}

// 	} else {
// 		if body.Tee == "1A" {
// 			teeList = []string{"10A", "1B", "10B"}
// 		} else if body.Tee == "10A" {
// 			teeList = []string{"1B", "10B", "1A"}
// 		} else if body.Tee == "1B" {
// 			teeList = []string{"10B", "1A", "10A"}
// 		} else {
// 			teeList = []string{"1A", "10A", "1B"}
// 		}
// 	}

// 	for index, tee := range teeList {
// 		t := teeTimeDate.Add((time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(turnLength)) * time.Duration(index))
// 		teeTime1B := strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())

// 		teeTimeLock := models.TeeInfo{
// 			TeeTime: teeTime1B,
// 			TeeType: tee,
// 		}
// 		listTeeTimeLock = append(listTeeTimeLock, teeTimeLock)
// 	}

// 	teeTimeSetting := models.LockTurn{
// 		TeeTimeLock:    listTeeTimeLock,
// 		BookingDate:    body.BookingDate,
// 		TurnTimeStatus: body.TurnTimeStatus,
// 		Tee:            body.Tee,
// 		CourseUid:      body.CourseUid,
// 		PartnerUid:     body.PartnerUid,
// 	}

// 	errC := teeTimeSetting.Create()
// 	if errC != nil {
// 		response_message.InternalServerError(c, errC.Error())
// 		return
// 	}

// 	okResponse(c, teeTimeSetting)
// }

func (_ *CLockTurn) GetLockTurn(c *gin.Context, prof models.CmsUser) {
	query := request.GetListLockTurn{}
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

	lockTurn := models.LockTurn{}

	if query.PartnerUid != "" {
		lockTurn.PartnerUid = query.PartnerUid
	}

	if query.CourseUid != "" {
		lockTurn.CourseUid = query.CourseUid
	}

	if query.BookingDate != "" {
		lockTurn.BookingDate = query.BookingDate
	}

	if query.TurnTimeStatus != "" {
		lockTurn.TurnTimeStatus = query.TurnTimeStatus
	}

	list, total, err := lockTurn.FindList(page)

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

func (_ *CLockTurn) DeleteLockTurn(c *gin.Context, prof models.CmsUser) {
	id := c.Param("id")
	idN, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		response_message.BadRequest(c, errors.New("id not valid").Error())
		return
	}

	lockTurn := models.LockTurn{}
	lockTurn.Id = idN
	errF := lockTurn.FindFirst()
	if errF != nil {
		response_message.InternalServerError(c, errF.Error())
		return
	}

	errDel := lockTurn.Delete()
	if errDel != nil {
		response_message.InternalServerError(c, errDel.Error())
		return
	}

	okRes(c)
}
