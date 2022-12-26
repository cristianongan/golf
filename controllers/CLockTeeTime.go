package controllers

import (
	"errors"
	"log"
	"start/config"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	"start/utils"
	"start/utils/response_message"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CLockTeeTime struct{}

func (_ *CLockTeeTime) CreateTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	body := request.CreateTeeTimeSettings{}
	if bindErr := c.ShouldBind(&body); bindErr != nil {
		badRequest(c, bindErr.Error())
		return
	}

	teeTimeSetting := models.LockTeeTime{
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		DateTime:       body.DateTime,
		TeeTime:        body.TeeTime,
		CurrentTeeTime: body.TeeTime,
		TeeType:        body.TeeType,
	}

	teeTimeRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:" + body.DateTime + "_" + body.CourseUid + "_"
	teeTimeRedisKey += body.TeeTime + "_" + body.TeeType

	key := datasources.GetRedisKeyTeeTimeLock(teeTimeRedisKey)
	_, errRedis := datasources.GetCache(key)

	teeTimeRedis := models.LockTeeTimeWithSlot{
		DateTime:       teeTimeSetting.DateTime,
		CourseUid:      teeTimeSetting.CourseUid,
		TeeTime:        teeTimeSetting.TeeTime,
		CurrentTeeTime: teeTimeSetting.TeeTime,
		TeeType:        teeTimeSetting.TeeType,
		TeeTimeStatus:  constants.TEE_TIME_LOCKED,
		Type:           constants.BOOKING_CMS,
	}

	if errRedis != nil {
		valueParse, _ := teeTimeRedis.Value()
		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 0); err != nil {

		}
	}

	okResponse(c, teeTimeSetting)
}
func (_ *CLockTeeTime) GetTeeTimeSettings(c *gin.Context, prof models.CmsUser) {
	query := request.GetListTeeTimeSettings{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	teeTimeSetting := models.LockTeeTime{}
	teeTimeSetting.TeeTime = query.TeeTime
	teeTimeSetting.TeeTimeStatus = query.TeeTimeStatus
	teeTimeSetting.DateTime = query.DateTime
	list := []models.LockTeeTimeWithSlot{}

	// get các teetime đang bị khóa ở redis
	listTeeTimeLockRedis := getTeeTimeLockRedis(query.CourseUid, query.DateTime)

	if query.RequestType == "TURN_TIME" {
		for _, teeTime := range listTeeTimeLockRedis {
			if teeTime.CurrentTeeTime != teeTime.TeeTime {
				list = append(list, teeTime)
			}
		}
	}

	if query.RequestType == "TEE_TIME" {
		for _, teeTime := range listTeeTimeLockRedis {
			if teeTime.CurrentTeeTime == teeTime.TeeTime || teeTime.TeeTime == "" {
				list = append(list, teeTime)
			}
		}
	}

	res := response.PageResponse{
		Total: int64(len(list)),
		Data:  list,
	}

	c.JSON(200, res)
}
func (_ *CLockTeeTime) LockTurn(body request.CreateLockTurn, c *gin.Context, prof models.CmsUser) error {
	db := datasources.GetDatabaseWithPartner(prof.PartnerUid)
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
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	weekday := strconv.Itoa(int(time.Now().Weekday() + 1))
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

		hour := t.Hour()
		minute := t.Minute()

		hourStr_ := strconv.Itoa(hour)
		if hour < 10 {
			hourStr_ = "0" + hourStr_
		}
		minuteStr := strconv.Itoa(minute)
		if minute < 10 {
			minuteStr = "0" + minuteStr
		}

		teeTime1B := hourStr_ + ":" + minuteStr

		lockTeeTime := models.LockTeeTime{
			PartnerUid:     body.PartnerUid,
			CourseUid:      body.CourseUid,
			TeeTime:        teeTime1B,
			TeeTimeStatus:  "LOCKED",
			DateTime:       body.BookingDate,
			CurrentTeeTime: body.TeeTime,
			TeeType:        data,
		}

		lockTeeTimeToRedis(lockTeeTime)
	}

	return nil
}
func (_ *CLockTeeTime) DeleteLockTurn(db *gorm.DB, teeTime string, bookingDate string, courseUid string) error {
	listTeeTimeLockRedis := getTeeTimeLockRedis(courseUid, bookingDate)

	for _, teeTimeR := range listTeeTimeLockRedis {
		if teeTimeR.CurrentTeeTime == teeTime {
			teeTimeRedisKey := config.GetEnvironmentName() + ":" + courseUid + "_" + bookingDate + "_" + teeTime + "_" + "1A"

			if err := datasources.DelCacheByKey(teeTimeRedisKey); err != nil {
				log.Println("DeleteLockTurn", err)
			}
		}
	}
	return nil
}

func (_ *CLockTeeTime) DeleteLockTeeTime(c *gin.Context, prof models.CmsUser) {
	query := request.DeleteLockRequest{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	list := []models.LockTeeTimeWithSlot{}

	// get các teetime đang bị khóa ở redis
	listTeeTimeLockRedis := getTeeTimeLockRedis(query.CourseUid, query.BookingDate)

	for _, teeTime := range listTeeTimeLockRedis {
		if teeTime.CurrentTeeTime == query.TeeTime {
			list = append(list, teeTime)
		}
	}

	for _, teeTime := range list {
		teeTimeRedisKey := config.GetEnvironmentName() + ":" + query.CourseUid + "_" + query.BookingDate + "_" + teeTime.TeeTime + "_" + teeTime.TeeType
		err := datasources.DelCacheByKey(teeTimeRedisKey)
		log.Print(err)
	}
	okRes(c)
}
