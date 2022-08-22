package controllers

import (
	"errors"
	"start/controllers/request"
	"start/controllers/response"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CLockTurn struct{}

func (_ *CLockTurn) CreateLockTurn(c *gin.Context, prof models.CmsUser) {
	body := request.CreateLockTurn{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	course := models.Course{}
	course.Uid = body.CourseUid
	errCourse := course.FindFirst()
	if errCourse != nil {
		response_message.InternalServerError(c, errCourse.Error())
		return
	}

	form := request.GetListBookingSettingForm{
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
	}

	cBookingSetting := CBookingSetting{}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(form)
	weekday := strconv.Itoa(int(time.Now().Weekday()))
	turnTimeH := 2
	turnLength := 0

	for _, data := range listSettingDetail {
		if strings.ContainsAny(data.Dow, weekday) {
			turnLength = data.TurnLength
			break
		}
	}

	listTeeTimeLock := models.ListTee{}
	teeTimeDate, _ := utils.ConvertHourToTime(body.TeeTime)

	if course.Hole == 18 {
		teeTimeLock1 := models.TeeInfo{
			TeeTime: body.TeeTime,
			TeeType: "1",
		}

		timeN := teeTimeDate.Add(time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(turnLength))
		teeTime10 := strconv.Itoa(timeN.Hour()) + ":" + strconv.Itoa(timeN.Minute())

		teeTimeLock10 := models.TeeInfo{
			TeeTime: teeTime10,
			TeeType: "10",
		}

		listTeeTimeLock = append(listTeeTimeLock, teeTimeLock1)
		listTeeTimeLock = append(listTeeTimeLock, teeTimeLock10)
	} else if course.Hole == 27 {
		teeTimeLock1A := models.TeeInfo{
			TeeTime: body.TeeTime,
			TeeType: "1A",
		}

		t := teeTimeDate.Add(time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(turnLength))
		teeTime1B := strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())

		teeTimeLock1B := models.TeeInfo{
			TeeTime: teeTime1B,
			TeeType: "1B",
		}

		t1 := teeTimeDate.Add((time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(turnLength)) * 2)
		teeTime1C := strconv.Itoa(t1.Hour()) + ":" + strconv.Itoa(t1.Minute())

		teeTimeLock1C := models.TeeInfo{
			TeeTime: teeTime1C,
			TeeType: "1C",
		}

		listTeeTimeLock = append(listTeeTimeLock, teeTimeLock1A)
		listTeeTimeLock = append(listTeeTimeLock, teeTimeLock1B)
		listTeeTimeLock = append(listTeeTimeLock, teeTimeLock1C)
	} else {
		// TODO 36 Holes
	}

	teeTimeSetting := models.LockTurn{
		TeeTimeLock:   listTeeTimeLock,
		BookingDate:   body.BookingDate,
		TeeTimeStatus: body.TeeTimeStatus,
		CourseUid:     body.CourseUid,
		PartnerUid:    body.PartnerUid,
	}

	errC := teeTimeSetting.Create()
	if errC != nil {
		response_message.InternalServerError(c, errC.Error())
		return
	}

	okResponse(c, teeTimeSetting)
}

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
