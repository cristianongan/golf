package controllers

import (
	"errors"
	"log"
	"net/http"
	"start/config"
	"start/constants"
	"start/controllers/request"
	"start/controllers/response"
	"start/datasources"
	"start/models"
	model_booking "start/models/booking"
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

	teeTimeRedisKey := getKeyTeeTimeLockRedis(body.DateTime, body.CourseUid, body.TeeTime, body.TeeType)

	key := datasources.GetRedisKeyTeeTimeLock(teeTimeRedisKey)
	_, errRedis := datasources.GetCache(key)

	teeTimeRedis := models.LockTeeTimeWithSlot{
		DateTime:       body.DateTime,
		PartnerUid:     body.PartnerUid,
		CourseUid:      body.CourseUid,
		TeeTime:        body.TeeTime,
		CurrentTeeTime: body.TeeTime,
		TeeType:        body.TeeType,
		TeeTimeStatus:  constants.TEE_TIME_LOCKED,
		Type:           constants.LOCK_CMS,
		Slot:           4,
		Note:           body.Note,
	}

	if errRedis != nil {
		valueParse, _ := teeTimeRedis.Value()
		if err := datasources.SetCache(teeTimeRedisKey, valueParse, 0); err != nil {

		}
	}

	okResponse(c, teeTimeRedis)
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
	listTeeTimeLockRedis := getTeeTimeLockRedis(query.CourseUid, query.DateTime, "")

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
		OnDate:     body.BookingDate,
	}

	cBookingSetting := CBookingSetting{}
	listSettingDetail, _, _ := cBookingSetting.GetSettingOnDate(db, form)
	bookingDateTime, _ := time.Parse(constants.DATE_FORMAT_1, body.BookingDate)
	weekday := strconv.Itoa(int(bookingDateTime.Weekday()))

	log.Println("LockTurn-weekday:", weekday)
	turnTimeH := 2
	bookSetting := model_booking.BookingSetting{}

	for _, data := range listSettingDetail {
		if strings.ContainsAny(data.Dow, weekday) {
			bookSetting = data
			break
		}
	}

	currentTeeTimeDate, _ := utils.ConvertHourToTime(body.TeeTime)
	teeList := []string{}

	if course.Hole == 18 {

		if body.TeeType == "1A" {
			teeList = []string{"1B"}
		} else if body.TeeType == "1B" {
			teeList = []string{"1C"}
		} else if body.TeeType == "1C" {
			teeList = []string{"1A"}
		}
	} else if course.Hole == 27 {

		if body.TeeType == "1A" {
			teeList = []string{"1B", "1C"}
		} else if body.TeeType == "1B" {
			teeList = []string{"1C", "1A"}
		} else if body.TeeType == "1C" {
			teeList = []string{"1A", "1B"}
		}

	}

	if len(teeList) == 0 {
		log.Println(errors.New("Không tìm thấy sân"))
	}

	timeParts := []response.TeeTimePartOTA{
		{
			IsHideTeePart: bookSetting.IsHideTeePart1,
			StartPart:     bookSetting.StartPart1,
			EndPart:       bookSetting.EndPart1,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart2,
			StartPart:     bookSetting.StartPart2,
			EndPart:       bookSetting.EndPart2,
		},
		{
			IsHideTeePart: bookSetting.IsHideTeePart3,
			StartPart:     bookSetting.StartPart3,
			EndPart:       bookSetting.EndPart3,
		},
	}

	index := 0
	teeTimeListLL := []string{}

	for _, part := range timeParts {
		if !part.IsHideTeePart {
			endTime, _ := utils.ConvertHourToTime(part.EndPart)
			teeTimeInit, _ := utils.ConvertHourToTime(part.StartPart)
			for {
				index += 1

				hour := teeTimeInit.Hour()
				minute := teeTimeInit.Minute()

				hourStr_ := strconv.Itoa(hour)
				if hour < 10 {
					hourStr_ = "0" + hourStr_
				}
				minuteStr := strconv.Itoa(minute)
				if minute < 10 {
					minuteStr = "0" + minuteStr
				}

				hourStr := hourStr_ + ":" + minuteStr

				teeTimeListLL = append(teeTimeListLL, hourStr)
				teeTimeInit = teeTimeInit.Add(time.Minute * time.Duration(bookSetting.TeeMinutes))

				if teeTimeInit.Unix() > endTime.Unix() {
					break
				}
			}
		}
	}

	for index, data := range teeList {

		t := currentTeeTimeDate.Add((time.Hour*time.Duration(turnTimeH) + time.Minute*time.Duration(bookSetting.TurnLength)) * time.Duration(index+1))

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

		if utils.Contains(teeTimeListLL, teeTime1B) {
			lockTeeTime := models.LockTeeTimeWithSlot{
				PartnerUid:     body.PartnerUid,
				CourseUid:      body.CourseUid,
				TeeTime:        teeTime1B,
				TeeTimeStatus:  "LOCKED",
				DateTime:       body.BookingDate,
				CurrentTeeTime: body.TeeTime,
				TeeType:        data,
				Type:           constants.LOCK_CMS,
			}

			lockTeeTimeToRedis(lockTeeTime)
		}
	}

	return nil
}
func (_ *CLockTeeTime) DeleteLockTurn(db *gorm.DB, teeTime string, bookingDate string, courseUid string) error {
	listTeeTimeLockRedis := getTeeTimeLockRedis(courseUid, bookingDate, "")

	for _, teeTimeR := range listTeeTimeLockRedis {
		if teeTimeR.CurrentTeeTime == teeTime {
			teeTimeRedisKey := getKeyTeeTimeLockRedis(teeTimeR.DateTime, teeTimeR.CourseUid, teeTimeR.TeeTime, teeTimeR.TeeType)
			err := datasources.DelCacheByKey(teeTimeRedisKey)

			log.Print(err)
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

	if query.Type == constants.LOCK_OTA {
		response_message.ErrorResponse(c, http.StatusBadRequest, "", "Unlock Fail", constants.ERROR_DELETE_LOCK_OTA)
		return
	}

	teeTimeRedisKey := getKeyTeeTimeLockRedis(query.BookingDate, query.CourseUid, query.TeeTime, query.TeeType+query.CourseType)
	err := datasources.DelCacheByKey(teeTimeRedisKey)
	log.Print(err)
	okRes(c)
}

func (_ *CLockTeeTime) DeleteAllRedisTeeTime(c *gin.Context, prof models.CmsUser) {
	query := request.DeleteRedis{}
	if err := c.Bind(&query); err != nil {
		response_message.BadRequest(c, err.Error())
		return
	}

	// Xóa tee time lock
	teeTimeLockRedisKey := config.GetEnvironmentName() + ":" + "tee_time_lock:"
	listKey, _ := datasources.GetAllKeysWith(teeTimeLockRedisKey)
	errTeeTimeLock := datasources.DelCacheByKey(listKey...)
	log.Print(errTeeTimeLock)

	// Xóa row_index
	teeTimeRowIndexRedisKey := config.GetEnvironmentName() + ":" + "tee_time_row_index:"
	listRowIndexKey, _ := datasources.GetAllKeysWith(teeTimeRowIndexRedisKey)
	errTeeTimeRowIndex := datasources.DelCacheByKey(listRowIndexKey...)
	log.Print(errTeeTimeRowIndex)

	// Xóa slot tee time
	teeTimeSlotEmptyRedisKey := config.GetEnvironmentName() + ":" + "tee_time_slot_empty" + "_"
	listTeeTimeSlotKey, _ := datasources.GetAllKeysWith(teeTimeSlotEmptyRedisKey)
	errTeeTimeSlot := datasources.DelCacheByKey(listTeeTimeSlotKey...)
	log.Print(errTeeTimeSlot)

	okRes(c)
}
