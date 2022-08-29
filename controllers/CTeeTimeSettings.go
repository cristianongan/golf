package controllers

import (
	"errors"
	"start/constants"
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

type CTeeTimeSettings struct{}

func (_ *CTeeTimeSettings) CreateTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	body := request.CreateTeeTimeSettings{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	teeTimeStatusList := []string{constants.TEE_TIME_LOCKED, constants.STATUS_DELETE, constants.TEE_TIME_UNLOCK}

	if !checkStringInArray(teeTimeStatusList, body.TeeTimeStatus) {
		response_message.BadRequest(c, "Tee Time Status incorrect")
		return
	}

	teeTimeSetting := models.LockTeeTime{
		TeeTime:        body.TeeTime,
		DateTime:       body.DateTime,
		CourseUid:      body.CourseUid,
		PartnerUid:     body.PartnerUid,
		CurrentTeeTime: body.TeeTime,
		TeeType:        body.TeeType,
	}

	errFind := teeTimeSetting.FindFirst()
	teeTimeSetting.TeeTimeStatus = body.TeeTimeStatus
	teeTimeSetting.Note = body.Note

	if errFind == nil {
		errC := teeTimeSetting.Update()
		if errC != nil {
			response_message.InternalServerError(c, errC.Error())
			return
		}
	} else {
		errC := teeTimeSetting.Create()
		if errC != nil {
			response_message.InternalServerError(c, errC.Error())
			return
		}
	}
	okResponse(c, teeTimeSetting)
}
func (_ *CTeeTimeSettings) GetTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	query := request.GetListTeeTimeSettings{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	teeTimeSetting := models.LockTeeTime{}

	if query.TeeTime != "" {
		teeTimeSetting.TeeTime = query.TeeTime
	}

	if query.TeeTimeStatus != "" {
		teeTimeSetting.TeeTimeStatus = query.TeeTimeStatus
	}

	if query.DateTime != "" {
		teeTimeSetting.DateTime = query.DateTime
	}

	list, total, err := teeTimeSetting.FindList(query.RequestType)

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
func (_ *CTeeTimeSettings) LockTurn(body request.CreateLockTurn, c *gin.Context, prof models.CmsUser) error {
	course := models.Course{}
	course.Uid = body.CourseUid
	errCourse := course.FindFirst()
	if errCourse != nil {
		return errCourse
	}

	form := request.GetListBookingSettingForm{
		CourseUid:  body.CourseUid,
		PartnerUid: body.PartnerUid,
	}

	cBookingSetting := CBookingSetting{}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(form)
	weekday := strconv.Itoa(int(time.Now().Weekday()))
	turnTimeH := 2
	endTime := ""
	turnLength := 0

	for _, data := range listSettingDetail {
		if strings.ContainsAny(data.Dow, weekday) {
			turnLength = data.TurnLength
			endTime = data.EndPart3
			break
		}
	}

	currentTeeTimeDate, _ := utils.ConvertHourToTime(body.TeeTime)
	endTimeDate, _ := utils.ConvertHourToTime(endTime)

	teeList := []string{}

	if course.Hole == 18 {

		if body.TeeType == "1" {
			teeList = []string{"10"}
		} else {
			teeList = []string{"1"}
		}
	} else if course.Hole == 27 {

		if body.TeeType == "1A" {
			teeList = []string{"1B", "1C"}
		} else if body.TeeType == "1B" {
			teeList = []string{"1C", "1A"}
		} else if body.TeeType == "1C" {
			teeList = []string{"1A", "1B"}
		}

	} else {
		if body.TeeType == "1A" {
			teeList = []string{"10A", "1B", "10B"}
		} else if body.TeeType == "10A" {
			teeList = []string{"1B", "10B", "1A"}
		} else if body.TeeType == "1B" {
			teeList = []string{"10B", "1A", "10A"}
		} else {
			teeList = []string{"1A", "10A", "1B"}
		}
	}

	if len(teeList) == 0 {
		return errors.New("Không tìm thấy sân")
	}

	for index, data := range teeList {

		t := currentTeeTimeDate.Add((time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(turnLength)) * time.Duration(index+1))

		if t.After(endTimeDate) {
			break
		}

		teeTime1B := strconv.Itoa(t.Hour()) + ":" + strconv.Itoa(t.Minute())

		lockTeeTime := models.LockTeeTime{
			PartnerUid:     body.PartnerUid,
			CourseUid:      body.CourseUid,
			TeeTime:        teeTime1B,
			TeeTimeStatus:  "LOCKED",
			DateTime:       body.BookingDate,
			CurrentTeeTime: body.TeeTime,
			TeeType:        data,
		}
		errC := lockTeeTime.Create()
		if errC != nil {
			return errC
		}
	}

	return nil
}
func (_ *CTeeTimeSettings) DeleteLockTurn(teeTime string, bookingDate string) error {
	lockTeeTime := models.LockTeeTime{
		CurrentTeeTime: teeTime,
		DateTime:       bookingDate,
	}
	list, _, _ := lockTeeTime.FindList("TURN_TIME")

	for _, data := range list {
		err := data.Delete()
		if err != nil {
			return err
		}
	}
	return nil
}
